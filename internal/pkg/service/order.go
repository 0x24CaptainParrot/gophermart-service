package service

import (
	"context"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
)

type OrderService struct {
	repo repository.Order
}

func NewOrderService(repo repository.Order) *OrderService {
	return &OrderService{repo: repo}
}

func (os *OrderService) CreateOrder(ctx context.Context, order models.Order) error {
	status, err := os.repo.CheckOrderStatus(ctx, order.Number, order.UserId)
	if err != nil {
		return err
	}
	switch status {
	case repository.StatusOwnedByUser:
		return repository.ErrAlreadyPostedByUser
	case repository.StatusOwnedByOther:
		return repository.ErrAlreadyExists
	}
	return os.repo.CreateOrder(ctx, order)
}

func (os *OrderService) ListOrders(ctx context.Context, id int) ([]models.Order, error) {
	return os.repo.ListOrders(ctx, id)
}

func (os *OrderService) UpdateOrderStatus(ctx context.Context, number int64, status string, accrual int) error {
	return os.repo.UpdateOrderStatus(ctx, number, status, accrual)
}
