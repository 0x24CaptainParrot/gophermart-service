package service

import (
	"context"
	"errors"
	"fmt"
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

type ResponseInfo struct {
	RespStatusCode int
}

func NewOrderService(repo repository.Order) *OrderService {
	return &OrderService{repo: repo}
}

func (os *OrderService) CreateOrder(ctx context.Context, order models.Order) (*ResponseInfo, error) {
	status, err := os.repo.CheckOrderStatus(ctx, order.Number, order.UserID)
	if err != nil {
		return nil, err
	}
	switch status {
	case repository.StatusOwnedByUser:
		return &ResponseInfo{RespStatusCode: http.StatusOK}, nil

	case repository.StatusOwnedByOther:
		return nil, &OrderServiceError{
			RespStatusCode: http.StatusConflict,
			ErrMsg:         repository.ErrAlreadyExists,
		}

	case repository.StatusNotExists:
		err := os.repo.CreateOrder(ctx, order)
		if err != nil {
			if errors.Is(err, repository.ErrAlreadyExists) {
				return nil, &OrderServiceError{
					RespStatusCode: http.StatusConflict,
					ErrMsg:         repository.ErrAlreadyExists,
				}
			}
			return nil, &OrderServiceError{
				RespStatusCode: http.StatusInternalServerError,
				ErrMsg:         err,
			}
		}
		return &ResponseInfo{RespStatusCode: http.StatusAccepted}, nil

	default:
		return nil, &OrderServiceError{
			RespStatusCode: http.StatusInternalServerError,
			ErrMsg:         fmt.Errorf("unexpected order status: %s", status),
		}
	}
}

func (os *OrderService) ListOrders(ctx context.Context, userID int) ([]models.Order, error) {
	return os.repo.ListOrders(ctx, userID)
}

func (os *OrderService) CheckOrderStatus(ctx context.Context, orderID int64, userID int) (string, error) {
	return os.repo.CheckOrderStatus(ctx, orderID, userID)
}
