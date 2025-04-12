package service

import (
	"context"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
)

type Order interface {
	CreateOrder(ctx context.Context, order models.Order) error
	ListOrders(ctx context.Context, id int) ([]models.Order, error)
}

// type CreateListOrders interface {
// 	CreateOrder(ctx context.Context, order models.Order) error
// 	ListOrders(ctx context.Context, id int) ([]models.Order, error)
// }

// type ProcessOrderWP interface {
// 	UpdateOrderStatus(ctx context.Context, number int64, status string, accrual int) error
// }

type Balance interface {
	AddLoyaltyPoints(ctx context.Context, userID int, orderID int64)
}

type Withdraw interface {
}

type Authorization interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GenerateToken(ctx context.Context, login, password string) (string, error)
	ParseToken(ctx context.Context, tokenGot string) (int, error)
}

type Service struct {
	Order         Order
	Authorization Authorization
	// ProcessOrderWP ProcessOrderWP
}

func NewService(repo repository.Authorization, repOrder repository.Order) *Service {
	return &Service{
		Authorization: NewAuthService(repo),
		Order:         NewOrderService(repOrder),
	}
}

// func NewWorkerPoolService() *WorkerPool {
// }
