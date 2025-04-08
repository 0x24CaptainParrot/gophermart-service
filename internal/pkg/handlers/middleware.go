package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/service"
)

type ctxKey string

const userIDKey ctxKey = "userID"

func GetUserID(r *http.Request) (int, bool) {
	idRaw := r.Context().Value(userIDKey)
	if idRaw == "" {
		return 0, false
	}
	id, ok := idRaw.(int)
	return id, ok
}

func AuthenticateMiddleware(authService service.Authorization) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var token string

			cookie, err := r.Cookie("Authorization")
			if err == nil {
				token = cookie.Value
			} else {
				authHeader := r.Header.Get("Authorization")
				if strings.Contains(authHeader, "Bearer ") {
					token = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if token == "" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := authService.ParseToken(r.Context(), token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), userIDKey, userID)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
