package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/0x24CaptainParrot/gophermart-service/internal/logger"
	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/service"
	"github.com/0x24CaptainParrot/gophermart-service/internal/utils"
	"github.com/go-chi/chi"
)

type OrderHandler struct {
	OrderService           service.Order
	OrderProcessingService service.OrderProcessing
}

func NewOrderHandler(order service.Order, processOrders service.OrderProcessing) *OrderHandler {
	return &OrderHandler{
		OrderService:           order,
		OrderProcessingService: processOrders,
	}
}

func (h *OrderHandler) OrderRoutes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.ProcessUserOrderHandler)
	r.Get("/", h.UserOrdersHandler)
	return r
}

func (h *OrderHandler) ProcessUserOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "user id is missing in context", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read body request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	num, err := strconv.ParseInt(string(body), 10, 64)
	if err != nil {
		http.Error(w, "invalid order number", http.StatusBadRequest)
		return
	}

	if !utils.IsValidOrderNum(num) {
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		return
	}

	order := models.Order{
		UserID: userID,
		Number: num,
		Status: "NEW",
	}

	ctx := r.Context()
	respInfo, err := h.OrderService.CreateOrder(ctx, order)
	if err != nil {
		var svcErr *service.OrderServiceError
		if errors.As(err, &svcErr) {
			http.Error(w, svcErr.Error(), svcErr.RespStatusCode)
			return
		}
		http.Error(w, "failed to return the server response", http.StatusInternalServerError)
		return
	}

	if respInfo.RespStatusCode == http.StatusAccepted {
		if err := h.OrderProcessingService.EnqueueOrder(r.Context(), order); err != nil {
			logger.Log.Sugar().Errorf("failed to enqueue order %d, err: %v", order.Number, err)
			http.Error(w, "failed to enqueue order", http.StatusInternalServerError)
			return
		}
	}

	w.WriteHeader(respInfo.RespStatusCode)
}

func (h *OrderHandler) UserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "user id is missing in context", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	orders, err := h.OrderService.ListOrders(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}
