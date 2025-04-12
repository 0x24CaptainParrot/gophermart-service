package repository

import (
	"context"
	"database/sql"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
)

type Authorization interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetUser(ctx context.Context, login, password string) (models.User, error)
}

type Order interface {
	CreateOrder(ctx context.Context, order models.Order) error
	ListOrders(ctx context.Context, userID int) ([]models.Order, error)
	CheckOrderStatus(ctx context.Context, orderID int64, userID int) (string, error)
	UpdateOrderStatus(ctx context.Context, number int64, status string, accrual int) error
}

type Balance interface {
	AddLoyaltyPoints(ctx context.Context, userID int, orderID int64)
}

type Repository struct {
	Authorization Authorization
	Order         Order
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
		Order:         NewOrderPostgres(db),
	}
}
