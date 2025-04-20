package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
)

type BalancePostgres struct {
	db *sql.DB
}

func NewBalancePostgres(db *sql.DB) *BalancePostgres {
	return &BalancePostgres{db: db}
}

const insertBalanceIfNotExists = `
				INSERT INTO balance (user_id, current, withdrawn) 
				VALUES ($1, 0, 0) 
				ON CONFLICT (user_id) DO NOTHING;`

const getUserBalance = `SELECT current, withdrawn FROM balance WHERE user_id = $1;`

var ErrNoPoints = errors.New("user has no loyalty points")

func (bp *BalancePostgres) DisplayUserBalance(ctx context.Context, userID int) (models.Balance, error) {
	if _, err := bp.db.ExecContext(ctx, insertBalanceIfNotExists, userID); err != nil {
		return models.Balance{}, fmt.Errorf("failed to insert balance row: %w", err)
	}

	var balance models.Balance
	row := bp.db.QueryRowContext(ctx, getUserBalance, userID)
	err := row.Scan(&balance.Current, &balance.Withdrawn)

	return balance, err
}

const insertWithdrawal = `INSERT INTO withdrawals (user_id, order_id, sum) VALUES ($1, $2, $3)`

const updateBalanceOnWithdraw = `
				UPDATE balance 
				SET current = current - $1, withdrawn = withdrawn + $1 
				WHERE user_id = $2 AND current >= $1`

var ErrInsufficientBalance = errors.New("insufficient loyalty points")

func (bp *BalancePostgres) WithdrawLoyaltyPoints(ctx context.Context, userID int, withdraw models.WithdrawRequest) error {
	tx, err := bp.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, updateBalanceOnWithdraw, withdraw.Sum, userID)
	if err != nil {
		return fmt.Errorf("failed to update balance: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrInsufficientBalance
	}

	_, err = tx.ExecContext(ctx, insertWithdrawal, userID, withdraw.Order, withdraw.Sum)
	if err != nil {
		return fmt.Errorf("failed to insert withdrawal: %w", err)
	}

	return tx.Commit()
}

const getWithdrawals = `SELECT order_id, sum, processed_at 
					FROM withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`

var ErrNoWithdrawals = errors.New("user has no withdrawals")

func (bp *BalancePostgres) DisplayWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	rows, err := bp.db.QueryContext(ctx, getWithdrawals, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	withdrawals := make([]models.Withdrawal, 0)
	for rows.Next() {
		var w models.Withdrawal
		timeStr := new(string)
		if err := rows.Scan(&w.Order, &w.Sum, timeStr); err != nil {
			return nil, err
		}
		if w.ProcessedAt, err = time.Parse(time.RFC3339, *timeStr); err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, w)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(withdrawals) == 0 {
		return nil, ErrNoWithdrawals
	}

	return withdrawals, nil
}
