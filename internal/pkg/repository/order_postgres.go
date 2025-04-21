package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type OrderPostgres struct {
	db *sql.DB
}

func NewOrderPostgres(db *sql.DB) *OrderPostgres {
	return &OrderPostgres{db: db}
}

var ErrAlreadyPostedByUser = errors.New("user already posted this order")

const createOrder = `INSERT INTO orders (user_id, number, status) VALUES ($1, $2, $3)`

func (op *OrderPostgres) CreateOrder(ctx context.Context, order models.Order) error {
	_, err := op.db.ExecContext(ctx, createOrder, order.UserID, order.Number, order.Status)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
		return ErrAlreadyPostedByUser
	}
	return nil
}

const getOrders = `SELECT number, status, accrual, uploaded_at 
					FROM orders WHERE user_id = $1 ORDER BY uploaded_at DESC`

func (op *OrderPostgres) ListOrders(ctx context.Context, userID int) ([]models.Order, error) {
	var orders []models.Order
	rows, err := op.db.QueryContext(ctx, getOrders, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		var acc sql.NullInt64
		var uploadedAt sql.NullTime
		if err := rows.Scan(&order.Number, &order.Status, &acc, &uploadedAt); err != nil {
			return nil, err
		}
		if acc.Valid {
			order.Accrual = acc.Int64
		}
		if uploadedAt.Valid {
			order.CreatedAt = uploadedAt.Time
		}
		orders = append(orders, order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return orders, nil
}

var ErrAlreadyExists = errors.New("order was already posted by another user")

const (
	StatusNotExists    = "NOT_EXISTS"
	StatusOwnedByUser  = "OWNED_BY_USER"
	StatusOwnedByOther = "OWNED_BY_OTHER"
)

const checkAvailability = `SELECT user_id FROM orders WHERE number = $1`

func (op *OrderPostgres) CheckOrderStatus(ctx context.Context, orderID int64, userID int) (string, error) {
	var dbUserID int
	err := op.db.QueryRowContext(ctx, checkAvailability, orderID).Scan(&dbUserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return StatusNotExists, nil
		}
		return "", err
	}

	if dbUserID == userID {
		return StatusOwnedByUser, nil
	}

	return StatusOwnedByOther, nil
}
