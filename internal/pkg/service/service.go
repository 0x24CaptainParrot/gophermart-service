package service

import (
	"context"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
)

type Order interface {
	CreateOrder(ctx context.Context, order models.Order) (*ResponseInfo, error)
	ListOrders(ctx context.Context, userID int) ([]models.Order, error)
	CheckOrderStatus(ctx context.Context, orderID int64, userID int) (string, error)
}

type OrderProcessing interface {
	StartProcessing(ctx context.Context, workers int)
	EnqueueOrder(ctx context.Context, order models.Order) error
	StopProcessing()
}

type Balance interface {
	DisplayUserBalance(ctx context.Context, userID int) (models.Balance, error)
	WithdrawLoyaltyPoints(ctx context.Context, userID int, withdraw models.WithdrawRequest) error
	DisplayWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error)
}

type Authorization interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GenerateToken(ctx context.Context, login, password string) (string, error)
	ParseToken(ctx context.Context, tokenGot string) (int, error)
}

type Service struct {
	Order           Order
	Authorization   Authorization
	Balance         Balance
	OrderProcessing OrderProcessing
}

func NewService(repos repository.Authorization, repOrder repository.Order, repBal repository.Balance, orderWp OrderProcessing) *Service {
	return &Service{
		Authorization:   NewAuthService(repos),
		Order:           NewOrderService(repOrder),
		Balance:         NewBalanceService(repBal),
		OrderProcessing: orderWp,
	}
}
