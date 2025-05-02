package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/0x24CaptainParrot/gophermart-service/internal/logger"
	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/service"
	"github.com/0x24CaptainParrot/gophermart-service/internal/utils"
)

type BalanceHandler struct {
	OrderService           service.Order
	BalanceService         service.Balance
	OrderProcessingService service.OrderProcessing
}

func NewBalanceHandler(order service.Order, balance service.Balance, processOrders service.OrderProcessing) *BalanceHandler {
	return &BalanceHandler{
		OrderService:           order,
		BalanceService:         balance,
		OrderProcessingService: processOrders,
	}
}

func (h *BalanceHandler) UserBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "user is is missing in context", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	balance, err := h.BalanceService.DisplayUserBalance(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

func (h *BalanceHandler) WithdrawLoyaltyPointsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "user id is missing in context", http.StatusUnauthorized)
		return
	}

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "invalid content-type", http.StatusUnsupportedMediaType)
		return
	}

	var withdrawInfo models.WithdrawRequest
	if err := json.NewDecoder(r.Body).Decode(&withdrawInfo); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !utils.IsValidOrderNum(withdrawInfo.Order) {
		http.Error(w, "invalid order number", http.StatusUnprocessableEntity)
		return
	}

	order := models.Order{
		UserID: userID,
		Number: withdrawInfo.Order,
		Status: "NEW",
	}

	respInfo, err := h.OrderService.CreateOrder(r.Context(), order)
	if err != nil {
		svcErr, ok := err.(*service.OrderServiceError)
		if !ok {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if svcErr.RespStatusCode == http.StatusConflict {
			http.Error(w, svcErr.Error(), http.StatusConflict)
			return
		}
		http.Error(w, svcErr.Error(), svcErr.RespStatusCode)
		return
	}

	if respInfo.RespStatusCode == http.StatusAccepted {
		if err := h.OrderProcessingService.EnqueueOrder(r.Context(), order); err != nil {
			logger.Log.Sugar().Errorf("failed to enqueue order %d, err: %v", order.Number, err)
			http.Error(w, "failed to enqueue order", http.StatusInternalServerError)
			return
		}
	}

	ctx := r.Context()
	if err := h.BalanceService.WithdrawLoyaltyPoints(ctx, userID, withdrawInfo); err != nil {
		if errors.Is(err, repository.ErrInsufficientBalance) {
			http.Error(w, err.Error(), http.StatusPaymentRequired)
			logger.Log.Sugar().Errorf("Withdraw failed: %v", err)
			return
		}
		logger.Log.Sugar().Errorf("Withdrawal failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) DisplayUserWithdrawalsHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "user id is missing in context", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	userWithdrawals, err := h.BalanceService.DisplayWithdrawals(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if len(userWithdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userWithdrawals)
}
