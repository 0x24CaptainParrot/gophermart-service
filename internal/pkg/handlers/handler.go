package handlers

import (
	"github.com/0x24CaptainParrot/gophermart-service/internal/config"
	"github.com/0x24CaptainParrot/gophermart-service/internal/logger"
	"github.com/0x24CaptainParrot/gophermart-service/internal/middleware"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/service"
	"github.com/go-chi/chi"
)

type Handler struct {
	Auth     *AuthHandler
	Orders   *OrderHandler
	Balance  *BalanceHandler
	services *service.Service
	cfg      *config.Config
}

func NewHandler(config *config.Config, service *service.Service) *Handler {
	return &Handler{
		Auth:     NewAuthHandler(service.Authorization),
		Orders:   NewOrderHandler(service.Order, service.OrderProcessing),
		Balance:  NewBalanceHandler(service.Order, service.Balance, service.OrderProcessing),
		services: service,
		cfg:      config,
	}
}

func (h *Handler) InitAPIRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(logger.LoggingReqResMiddleware(logger.Log))
	r.Use(middleware.CompressGzipMiddleware())

	r.Route("/api", func(r chi.Router) {
		r.Route("/user", func(r chi.Router) {
			r.Post("/register", h.Auth.RegisterUserHandler)
			r.Post("/login", h.Auth.LoginHandler)

			r.Group(func(r chi.Router) {
				r.Use(AuthenticateMiddleware(h.services.Authorization))
				r.Post("/orders", h.Orders.ProcessUserOrderHandler)
				r.Get("/orders", h.Orders.UserOrdersHandler)
				r.Route("/balance", func(r chi.Router) {
					r.Get("/", h.Balance.UserBalanceHandler)
					r.Post("/withdraw", h.Balance.WithdrawLoyaltyPointsHandler)
				})
				r.Get("/withdrawals", h.Balance.DisplayUserWithdrawalsHandler)
			})
		})
	})

	return r
}
