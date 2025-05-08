package handlers

import (
	"github.com/0x24CaptainParrot/gophermart-service/internal/config"
	"github.com/0x24CaptainParrot/gophermart-service/internal/logger"
	"github.com/0x24CaptainParrot/gophermart-service/internal/middleware"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/service"
	"github.com/go-chi/chi"
)

type Handler struct {
	AuthHandler    *AuthHandler
	OrdersHandler  *OrderHandler
	BalanceHandler *BalanceHandler
	services       *service.Service
	cfg            *config.Config
}

func NewHandler(config *config.Config, service *service.Service) *Handler {
	return &Handler{
		AuthHandler:   NewAuthHandler(service.Authorization),
		OrdersHandler: NewOrderHandler(service.Order, service.OrderProcessing),
		BalanceHandler: NewBalanceHandler(
			WithOrderService(service.Order),
			WithBalanceService(service.Balance),
			WithOrderProcessingService(service.OrderProcessing),
		),
		services: service,
		cfg:      config,
	}
}

func (h *Handler) InitAPIRoutes() *chi.Mux {
	r := chi.NewRouter()
	r.Use(logger.LoggingReqResMiddleware(logger.Log))
	r.Use(middleware.CompressGzipMiddleware())

	r.Mount("/api/user", h.userRouter())

	return r
}

func (h *Handler) userRouter() chi.Router {
	r := chi.NewRouter()
	r.Mount("/", h.AuthHandler.AuthRoutes())

	r.Group(func(r chi.Router) {
		r.Use(AuthenticateMiddleware(h.services.Authorization))

		r.Mount("/orders", h.OrdersHandler.OrderRoutes())
		r.Mount("/balance", h.BalanceHandler.BalanceRoutes())
		r.Mount("/withdrawals", h.BalanceHandler.WithdrawalsRoutes())
	})

	return r
}
