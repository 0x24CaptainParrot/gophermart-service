package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type AuthPostgres struct {
	db *sql.DB
}

func NewAuthPostgres(db *sql.DB) *AuthPostgres {
	return &AuthPostgres{db: db}
}

var ErrUserExists = errors.New("user already exists")

const createUser = `INSERT INTO users (login, password_hash) VALUES ($1, $2) RETURNING id`

func (ap *AuthPostgres) CreateUser(ctx context.Context, user models.User) (int, error) {
	var userID int

	row := ap.db.QueryRowContext(ctx, createUser, user.Login, user.Password)
	err := row.Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return 0, ErrUserExists
		}
		return 0, err
	}
	return userID, nil
}

const getUser = `SELECT id FROM users WHERE login=$1 AND password_hash=$2`

func (ap *AuthPostgres) GetUser(ctx context.Context, login, password string) (models.User, error) {
	var user models.User

	row := ap.db.QueryRowContext(ctx, getUser, login, password)
	err := row.Scan(&user.ID)

	return user, err
}
