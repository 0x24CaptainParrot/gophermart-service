package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/0x24CaptainParrot/gophermart-service/internal/mocks"
	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestRegisterUserHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuthorization(ctrl)
	handler := NewAuthHandler(mockAuth)

	tests := []struct {
		name           string
		input          models.User
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "successful registration",
			input: models.User{
				Login:    "user",
				Password: "password",
			},
			mockSetup: func() {
				mockAuth.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(1, nil)
				mockAuth.EXPECT().GenerateToken(gomock.Any(), "user", "password").Return("token", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "user already exists",
			input: models.User{
				Login:    "alreadyExists",
				Password: "password",
			},
			mockSetup: func() {
				mockAuth.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(0, repository.ErrUserExists)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "internal error on create",
			input: models.User{
				Login:    "userError",
				Password: "password",
			},
			mockSetup: func() {
				mockAuth.EXPECT().CreateUser(gomock.Any(), gomock.Any()).Return(0, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "invalid json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			var req *http.Request
			if tt.name == "invalid json" {
				req = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer([]byte("{badJson}")))
			} else {
				body, _ := json.Marshal(tt.input)
				req = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(body))
			}

			rec := httptest.NewRecorder()
			handler.RegisterUserHandler(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestLoginHandler(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuth := mocks.NewMockAuthorization(ctrl)
	handler := NewAuthHandler(mockAuth)

	tests := []struct {
		name           string
		input          models.User
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "successful login",
			input: models.User{
				Login:    "user",
				Password: "password",
			},
			mockSetup: func() {
				mockAuth.EXPECT().GenerateToken(gomock.Any(), "user", "password").Return("token", nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid credentials",
			input: models.User{
				Login:    "badUser",
				Password: "wrong",
			},
			mockSetup: func() {
				mockAuth.EXPECT().GenerateToken(gomock.Any(), "badUser", "wrong").Return("", errors.New("unauthorized"))
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "invalid json",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			var req *http.Request
			if tt.name == "invalid json" {
				req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer([]byte("badJson")))
			} else {
				body, _ := json.Marshal(tt.input)
				req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
			}

			rec := httptest.NewRecorder()
			handler.LoginHandler(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}
