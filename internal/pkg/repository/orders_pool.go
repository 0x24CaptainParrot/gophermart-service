package repository

import (
	"context"
	"fmt"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkerPoolRepo struct {
	pool *pgxpool.Pool
}

func NewWorkerPoolRepo(pool *pgxpool.Pool) *WorkerPoolRepo {
	return &WorkerPoolRepo{pool: pool}
}

const (
	updateOrderWithAccrual = `
		WITH updated_order AS (
			UPDATE orders 
			SET status = $2,
				accrual = $3,
				updated_at = NOW() 
			WHERE number = $1 
			RETURNING user_id, accrual
		)
		UPDATE balance 
		SET current = current + updated_order.accrual 
		FROM updated_order 
		WHERE balance.user_id = updated_order.user_id;`
)

func (r *WorkerPoolRepo) UpdateOrderAndBalance(ctx context.Context, order models.Order, accrual float64) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, updateOrderWithAccrual, order.Number, order.Status, accrual)
	if err != nil {
		return fmt.Errorf("failed to update order and balance: %w", err)
	}

	return tx.Commit(ctx)
}

const (
	getPendingOrders = `
	SELECT user_id, number, status 
		FROM orders WHERE status IN ('NEW', 'PROCESSING') 
	ORDER BY uploaded_at ASC LIMIT $1`
)

func (r *WorkerPoolRepo) GetPendingOrders(ctx context.Context, limit int) ([]models.Order, error) {
	rows, err := r.pool.Query(ctx, getPendingOrders, limit)
	if err != nil {
		return nil, err
	}

	orders := make([]models.Order, 0, limit)
	for rows.Next() {
		var order models.Order
		err := rows.Scan(&order.UserID, &order.Number, &order.Status)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}

	return orders, nil
}
