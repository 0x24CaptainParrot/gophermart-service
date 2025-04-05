package handlers

import (
	"net/http"

	"github.com/go-chi/chi"
)

type Handler struct {
	// services
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) InitApiRoutes() *chi.Mux {
	r := chi.NewRouter()

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", Adapter(h.RegisterUserHandler))
			r.Post("/login", Adapter(h.LoginHandler))
			// r.Post("/orders")
			// r.Get("/orders")
			// r.Route("/balance", func(r chi.Router) {
			// 	r.Get("/")
			// 	r.Post("/withdraw")
			// })
			// r.Get("/withdrawals")
		})
	})

	return r
}

type ApiFunc func(w http.ResponseWriter, r *http.Request) error

func Adapter(f ApiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
