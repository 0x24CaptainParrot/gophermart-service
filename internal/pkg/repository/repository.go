package repository

import (
	"context"
	"database/sql"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
)

type Authorization interface {
	CreateUser(ctx context.Context, user models.User) (int, error)
	GetUser(ctx context.Context, login, password string) (models.User, error)
}

type Repository struct {
	Authorization Authorization
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
	}
}
