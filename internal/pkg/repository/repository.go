package repository

import (
	"context"
	"database/sql"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
)

type AuthorizationRepository interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetUser(ctx context.Context, login, password string) (models.User, error)
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order models.Order) error
	ListOrders(ctx context.Context, userID int) ([]models.Order, error)
	CheckOrderStatus(ctx context.Context, orderID int64, userID int) (string, error)
}

type BalanceRepository interface {
	DisplayUserBalance(ctx context.Context, userID int) (models.Balance, error)
	WithdrawLoyaltyPoints(ctx context.Context, userID int, withdraw models.WithdrawRequest) error
	DisplayWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error)
}

type ProcessOrderWP interface {
	Start(ctx context.Context, workers int)
	AddOrder(ctx context.Context, order models.Order) error
	Stop()
}

type WorkerPoolRepository interface {
	UpdateOrderAndBalance(ctx context.Context, order models.Order, accrual float64) error
	GetPendingOrders(ctx context.Context, limit int) ([]models.Order, error)
	LockAndGetOrderStatus(ctx context.Context, orderNumber int64) (string, error)
	InsertMissingOrder(ctx context.Context, orderNumber int64) error
}

type Repository struct {
	Authorization AuthorizationRepository
	Order         OrderRepository
	Balance       BalanceRepository
	WPRepository  WorkerPoolRepository
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
		Order:         NewOrderPostgres(db),
		Balance:       NewBalancePostgres(db),
	}
}
