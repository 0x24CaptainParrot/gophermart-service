package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/0x24CaptainParrot/gophermart-service/internal/mocks"
	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func addUserToContext(r *http.Request, userID int) *http.Request {
	ctx := context.WithValue(r.Context(), userIDKey, userID)
	return r.WithContext(ctx)
}

func TestProcessUserOrderHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockOrder(ctrl)
	mockOrderProcessor := mocks.NewMockOrderProcessing(ctrl)
	handler := NewOrderHandler(mockOrder, mockOrderProcessor)

	validOrder := int64(79927398713)

	tests := []struct {
		name           string
		body           string
		mockSetup      func()
		contextUserID  int
		expectedStatus int
	}{
		{
			name:          "valid new order",
			body:          strconv.FormatInt(validOrder, 10),
			contextUserID: 123,
			mockSetup: func() {
				mockOrder.EXPECT().
					CreateOrder(gomock.Any(), gomock.Any()).
					Return(&service.ResponseInfo{RespStatusCode: http.StatusAccepted}, nil)
				mockOrderProcessor.EXPECT().
					EnqueueOrder(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			expectedStatus: http.StatusAccepted,
		},
		{
			name:          "order already posted by user",
			body:          strconv.FormatInt(validOrder, 10),
			contextUserID: 123,
			mockSetup: func() {
				mockOrder.EXPECT().
					CreateOrder(gomock.Any(), gomock.Any()).
					Return(&service.ResponseInfo{RespStatusCode: http.StatusOK}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "order already posted by another user",
			body:          strconv.FormatInt(validOrder, 10),
			contextUserID: 123,
			mockSetup: func() {
				mockOrder.EXPECT().
					CreateOrder(gomock.Any(), gomock.Any()).
					Return(nil, &service.OrderServiceError{
						RespStatusCode: http.StatusConflict,
						ErrMsg:         errors.New("already exists"),
					})
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "invalid order number format",
			body:           "invalidOrder123",
			contextUserID:  123,
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid luhn",
			body:           "123456789",
			contextUserID:  123,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:           "missing user id",
			body:           strconv.FormatInt(validOrder, 10),
			contextUserID:  0,
			mockSetup:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", bytes.NewBuffer([]byte(tt.body)))
			if tt.contextUserID != 0 {
				req = addUserToContext(req, tt.contextUserID)
			}

			rec := httptest.NewRecorder()
			handler.ProcessUserOrderHandler(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestUserOrdersHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockOrder := mocks.NewMockOrder(ctrl)
	handler := NewOrderHandler(mockOrder, nil)

	tests := []struct {
		name           string
		contextUserID  int
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:          "user has orders",
			contextUserID: 123,
			mockSetup: func() {
				mockOrder.EXPECT().
					ListOrders(gomock.Any(), 123).
					Return([]models.Order{{Number: 123456}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "no orders",
			contextUserID: 123,
			mockSetup: func() {
				mockOrder.EXPECT().
					ListOrders(gomock.Any(), 123).
					Return([]models.Order{}, nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:          "sql.ErrNoRows",
			contextUserID: 123,
			mockSetup: func() {
				mockOrder.EXPECT().
					ListOrders(gomock.Any(), 123).
					Return(nil, sql.ErrNoRows)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:          "internal error",
			contextUserID: 123,
			mockSetup: func() {
				mockOrder.EXPECT().
					ListOrders(gomock.Any(), 123).
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
			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			if tt.contextUserID != 0 {
				req = addUserToContext(req, tt.contextUserID)
			}

			rec := httptest.NewRecorder()
			handler.UserOrdersHandler(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
