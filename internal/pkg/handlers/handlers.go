package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
	"github.com/0x24CaptainParrot/gophermart-service/internal/utils"
)

func (h *Handler) ProcessUserOrderHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "invalid content-type", http.StatusUnsupportedMediaType)
		return
	}

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
	if !utils.IsValidOrderNumberLuhn(num) {
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		return
	}

	order := models.Order{
		UserId: userID,
		Number: num,
		Status: "NEW",
	}

	ctx := r.Context()
	if err := h.services.Order.CreateOrder(ctx, order); err != nil {
		switch true {
		case errors.Is(err, repository.ErrAlreadyPostedByUser):
			http.Error(w, err.Error(), http.StatusOK)
			return
		case errors.Is(err, repository.ErrAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)
			return
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// go h.AddLoyaltyPoints(ctx, order.UserId, order.Number)

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) UserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "invalid content type", http.StatusUnsupportedMediaType)
		return
	}

	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "user id is missing in context", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	orders, err := h.services.Order.ListOrders(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			http.Error(w, "user has no orders", http.StatusNoContent)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

func (h *Handler) AddLoyaltyPoints(ctx context.Context, UserID int, orderID int64) {
	var order struct {
		Order   string `json:"order"`
		Status  string `json:"status"`
		Accrual int    `json:"accrual"`
	}
	client := &http.Client{}
	resp, err := client.Get(fmt.Sprintf("%s/api/orders/%d", h.cfg.AccrualAddr, orderID))
	if err != nil {
		log.Println("failed to fetch")
		return
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		log.Fatal()
		return
	}

	// if err := h.services.Balance.AddLoyaltyPoints(ctx, UserID); err != nil {
	// }
}
