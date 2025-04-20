package service

import (
	"context"
	"net/http"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
)

type OrderService struct {
	repo repository.Order
}

type OrderServiceError struct {
	RespStatusCode int
	ErrMsg         error
}

func (oe *OrderServiceError) Error() string {
	return oe.ErrMsg.Error()
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
		return &OrderServiceError{
			RespStatusCode: http.StatusOK,
			ErrMsg:         repository.ErrAlreadyPostedByUser,
		}
	case repository.StatusOwnedByOther:
		return &OrderServiceError{
			RespStatusCode: http.StatusConflict,
			ErrMsg:         repository.ErrAlreadyExists,
		}
	}
	return os.repo.CreateOrder(ctx, order)
}

func (os *OrderService) ListOrders(ctx context.Context, userID int) ([]models.Order, error) {
	return os.repo.ListOrders(ctx, userID)
}

func (os *OrderService) CheckOrderStatus(ctx context.Context, orderID int64, userID int) (string, error) {
	return os.repo.CheckOrderStatus(ctx, orderID, userID)
}
