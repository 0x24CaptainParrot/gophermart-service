package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/0x24CaptainParrot/gophermart-service/internal/logger"
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
		),
		insert_balance AS (
			INSERT INTO balance (user_id, current, withdrawn) 
			SELECT user_id, 0, 0 FROM updated_order 
			ON CONFLICT (user_id) DO NOTHING
		)
		UPDATE balance 
		SET current = current + updated_order.accrual 
		FROM updated_order 
		WHERE balance.user_id = updated_order.user_id;`
)

func (r *WorkerPoolRepo) UpdateOrderAndBalance(ctx context.Context, order models.Order, accrual float64) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var locked bool
	if err := tx.QueryRow(ctx, `SELECT pg_try_advisory_xact_lock($1)`, order.Number).Scan(&locked); err != nil {
		return fmt.Errorf("failed to acquire advisory lock: %w", err)
	}
	if !locked {
		return fmt.Errorf("order %d is already locked", order.Number)
	}

	tag, err := tx.Exec(ctx, updateOrderWithAccrual, order.Number, order.Status, accrual)
	if err != nil {
		return fmt.Errorf("failed to update order and balance: %w", err)
	}

	if tag.RowsAffected() == 0 {
		logger.Log.Sugar().Infof("No rows updated: maybe order %d already in correct state", order.Number)
	}

	return tx.Commit(ctx)
}

const (
	getPendingOrders = `
	SELECT user_id, number, status 
	FROM orders 
	WHERE status IN ('NEW', 'PROCESSING') 
	AND pg_try_advisory_xact_lock(number) 
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

const lockGetOrderStatus = `SELECT status FROM orders WHERE number = $1 FOR UPDATE SKIP LOCKED`

var ErrOrderNotFound = errors.New("order was not found")

func (r *WorkerPoolRepo) LockAndGetOrderStatus(ctx context.Context, orderNumber int64) (string, error) {
	var status string
	err := r.pool.QueryRow(ctx, lockGetOrderStatus, orderNumber).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrOrderNotFound
		}
		return "", fmt.Errorf("failed to lock and get order status: %w", err)
	}
	return status, nil
}

const insertMissingOrder = `
				INSERT INTO orders (user_id, number, status) 
				VALUES (0, $1, 'NEW') 
				ON CONFLICT (number) DO NOTHING`

func (r *WorkerPoolRepo) InsertMissingOrder(ctx context.Context, orderNumber int64) error {
	_, err := r.pool.Exec(ctx, insertMissingOrder, orderNumber)
	if err != nil {
		return fmt.Errorf("failed to insert missing order: %w", err)
	}
	return nil
}
