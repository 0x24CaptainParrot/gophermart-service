package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/service"
)

type AuthHandler struct {
	AuthService service.Authorization
}

func NewAuthHandler(auth service.Authorization) *AuthHandler {
	return &AuthHandler{AuthService: auth}
}

// sign up
func (h *AuthHandler) RegisterUserHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	id, err := h.AuthService.CreateUser(ctx, user)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			http.Error(w, "user with the given login already exists", http.StatusConflict)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	token, err := h.AuthService.GenerateToken(ctx, user.Login, user.Password)
	if err != nil {
		log.Println(err)
		http.Error(w, "failed to generate token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"id": id})
}

// sign in
func (h *AuthHandler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	token, err := h.AuthService.GenerateToken(ctx, user.Login, user.Password)
	if err != nil {
		http.Error(w, "invalid login/password", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "Authorization",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{"token": token})
}
