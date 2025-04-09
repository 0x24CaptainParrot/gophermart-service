package handlers

import (
	"net/http"

	"github.com/0x24CaptainParrot/gophermart-service/internal/logger"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/service"
	"github.com/go-chi/chi"
)

type Handler struct {
	services *service.Service
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		services: service,
	}
}

func (h *Handler) InitAPIRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(logger.LoggingReqResMiddleware(logger.Log))

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", h.RegisterUserHandler)
			r.Post("/login", h.LoginHandler)

			r.Group(func(r chi.Router) {
				r.Use(AuthenticateMiddleware(h.services.Authorization))
				// r.Post("/orders")
				// r.Get("/orders")
				// r.Route("/balance", func(r chi.Router) {
				// 	r.Get("/")
				// 	r.Post("/withdraw")
				// })
				// r.Get("/withdrawals")
			})
		})
	})

	return r
}

type APIFunc func(w http.ResponseWriter, r *http.Request) error

func Adapter(f APIFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
