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
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/service"
)

func (h *Handler) ProcessUserOrderHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "user id is missing in context", http.StatusUnauthorized)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "invalid content-type", http.StatusUnsupportedMediaType)
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

	order := models.Order{
		UserId: userID,
		Number: num,
		Status: "NEW",
	}

	ctx := r.Context()
	if err := h.services.Order.CreateOrder(ctx, order); err != nil {
		respOrderErr, ok := err.(*service.OrderServiceError)
		if !ok {
			http.Error(w, "failed to return the server response", http.StatusInternalServerError)
			return
		}
		http.Error(w, respOrderErr.Error(), respOrderErr.RespStatusCode)
		return
	}

	if err := h.services.OrderProcessing.EnqueueOrder(r.Context(), order); err != nil {
		http.Error(w, "failed to enqueue order", http.StatusInternalServerError)
		logger.Log.Sugar().Errorf("failed to enqueue order %d, err: %v", order.Number, err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *Handler) UserOrdersHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "user id is missing in context", http.StatusUnauthorized)
		return
	}

	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "invalid content type", http.StatusUnsupportedMediaType)
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

func (h *Handler) UserBalanceHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r)
	if !ok {
		http.Error(w, "user is is missing in context", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	balance, err := h.services.Balance.DisplayUserBalance(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balance)
}

func (h *Handler) WithdrawLoyaltyPointsHandler(w http.ResponseWriter, r *http.Request) {
	userId, ok := GetUserID(r)
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

	order := models.Order{
		UserId: userId,
		Number: withdrawInfo.Order,
		Status: "NEW",
	}

	if err := h.services.Order.CreateOrder(r.Context(), order); err != nil {
		respOrderErr, ok := err.(*service.OrderServiceError)
		if !ok {
			http.Error(w, "failed to return the server response", http.StatusInternalServerError)
			return
		}
		http.Error(w, respOrderErr.Error(), respOrderErr.RespStatusCode)
		return
	}

	ctx := r.Context()
	if err := h.services.Balance.WithdrawLoyaltyPoints(ctx, userId, withdrawInfo); err != nil {
		if errors.Is(err, repository.ErrInsufficientBalance) {
			http.Error(w, err.Error(), http.StatusPaymentRequired)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DisplayUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	userId, ok := GetUserID(r)
	if !ok {
		http.Error(w, "user id is missing in context", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	userWithdrawals, err := h.services.Balance.DisplayWithdrawals(ctx, userId)
	if err != nil {
		if errors.Is(err, repository.ErrNoWithdrawals) {
			http.Error(w, err.Error(), http.StatusNoContent)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userWithdrawals)
}
