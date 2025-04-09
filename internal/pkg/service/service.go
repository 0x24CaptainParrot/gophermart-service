package service

import (
	"context"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
)

type Order interface {
}

type Balance interface {
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
}

func NewService(repo repository.Authorization) *Service {
	return &Service{
		Authorization: NewAuthService(repo),
	}
}
