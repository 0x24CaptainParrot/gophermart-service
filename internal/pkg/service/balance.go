package service

import (
	"context"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
)

type Balance interface {
	DisplayUserBalance(ctx context.Context, userID int) (models.Balance, error)
	WithdrawLoyaltyPoints(ctx context.Context, userID int, withdraw models.WithdrawRequest) error
	DisplayWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error)
}

type BalanceService struct {
	repo repository.BalanceRepository
}

func NewBalanceService(repo repository.BalanceRepository) *BalanceService {
	return &BalanceService{repo: repo}
}

func (bs *BalanceService) DisplayUserBalance(ctx context.Context, userID int) (models.Balance, error) {
	return bs.repo.DisplayUserBalance(ctx, userID)
}

func (bs *BalanceService) WithdrawLoyaltyPoints(ctx context.Context, userID int, withdraw models.WithdrawRequest) error {
	return bs.repo.WithdrawLoyaltyPoints(ctx, userID, withdraw)
}

func (bs *BalanceService) DisplayWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	return bs.repo.DisplayWithdrawals(ctx, userID)
}
