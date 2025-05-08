package handlers

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0x24CaptainParrot/gophermart-service/internal/mocks"
	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUserBalanceHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBalance := mocks.NewMockBalance(ctrl)
	handler := NewBalanceHandler(WithBalanceService(mockBalance))

	tests := []struct {
		name           string
		contextUserID  int
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:          "authorized and success",
			contextUserID: 123,
			mockSetup: func() {
				mockBalance.EXPECT().
					DisplayUserBalance(gomock.Any(), 123).
					Return(models.Balance{Current: 100, Withdrawn: 50}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user id",
			contextUserID:  0,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:          "internal error",
			contextUserID: 123,
			mockSetup: func() {
				mockBalance.EXPECT().
					DisplayUserBalance(gomock.Any(), 123).
					Return(models.Balance{}, errors.New("db failure"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/api/user/balance", nil)
			if tt.contextUserID != 0 {
				req = addUserToContext(req, tt.contextUserID)
			}

			rec := httptest.NewRecorder()
			handler.UserBalanceHandler(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestWithdrawLoyaltyPointsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBalance := mocks.NewMockBalance(ctrl)
	mockOrder := mocks.NewMockOrder(ctrl)
	mockOrderProcessor := mocks.NewMockOrderProcessing(ctrl)

	handler := NewBalanceHandler(
		WithBalanceService(mockBalance),
		WithOrderService(mockOrder),
		WithOrderProcessingService(mockOrderProcessor),
	)

	body := `{"order": "79927398713", "sum": 50}`

	tests := []struct {
		name           string
		body           string
		contentType    string
		contextUserID  int
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:          "success",
			body:          body,
			contentType:   "application/json",
			contextUserID: 123,
			mockSetup: func() {
				order := models.Order{UserID: 123, Number: 79927398713, Status: "NEW"}
				mockOrder.EXPECT().
					CreateOrder(gomock.Any(), order).
					Return(&service.ResponseInfo{RespStatusCode: http.StatusAccepted}, nil)
				mockOrderProcessor.EXPECT().
					EnqueueOrder(gomock.Any(), order).
					Return(nil)
				mockBalance.EXPECT().
					WithdrawLoyaltyPoints(gomock.Any(), 123, gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing user id",
			contextUserID:  0,
			contentType:    "application/json",
			body:           body,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid json",
			contextUserID:  123,
			contentType:    "application/json",
			body:           `{invalid json`,
			mockSetup:      func() {},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid content-type",
			contextUserID:  123,
			contentType:    "text/plain",
			body:           body,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnsupportedMediaType,
		},
		{
			name:           "invalid order number",
			contextUserID:  123,
			contentType:    "application/json",
			body:           `{"order": "12345", "sum": 10}`,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:          "insufficient balance",
			contextUserID: 123,
			contentType:   "application/json",
			body:          body,
			mockSetup: func() {
				order := models.Order{UserID: 123, Number: 79927398713, Status: "NEW"}
				mockOrder.EXPECT().
					CreateOrder(gomock.Any(), order).
					Return(&service.ResponseInfo{RespStatusCode: http.StatusOK}, nil)
				mockBalance.EXPECT().
					WithdrawLoyaltyPoints(gomock.Any(), 123, gomock.Any()).
					Return(repository.ErrInsufficientBalance)
			},
			expectedStatus: http.StatusPaymentRequired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBuffer([]byte(tt.body)))
			req.Header.Set("Content-Type", tt.contentType)
			if tt.contextUserID != 0 {
				req = addUserToContext(req, tt.contextUserID)
			}

			rec := httptest.NewRecorder()
			handler.WithdrawLoyaltyPointsHandler(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestDisplayUserWithdrawalsHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockBalance := mocks.NewMockBalance(ctrl)
	handler := NewBalanceHandler(WithBalanceService(mockBalance))

	tests := []struct {
		name           string
		contextUserID  int
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:          "success with data",
			contextUserID: 123,
			mockSetup: func() {
				mockBalance.EXPECT().
					DisplayWithdrawals(gomock.Any(), 123).
					Return([]models.Withdrawal{{Order: 12345, Sum: 50}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "no withdrawals",
			contextUserID: 123,
			mockSetup: func() {
				mockBalance.EXPECT().
					DisplayWithdrawals(gomock.Any(), 123).
					Return([]models.Withdrawal{}, nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:          "internal error",
			contextUserID: 123,
			mockSetup: func() {
				mockBalance.EXPECT().
					DisplayWithdrawals(gomock.Any(), 123).
					Return(nil, errors.New("db failure"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "missing user id",
			contextUserID:  0,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodGet, "/api/user/withdrawals", nil)
			if tt.contextUserID != 0 {
				req = addUserToContext(req, tt.contextUserID)
			}

			rec := httptest.NewRecorder()
			handler.DisplayUserWithdrawalsHandler(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
