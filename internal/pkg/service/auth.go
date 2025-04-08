package service

import (
	"context"
	"crypto/sha256"
	"fmt"
	"time"

	"github.com/0x24CaptainParrot/gophermart-service/internal/models"
	"github.com/0x24CaptainParrot/gophermart-service/internal/pkg/repository"
	"github.com/golang-jwt/jwt/v4"
)

const (
	TOKEN_EXP  = time.Hour * 6
	SECRET_KEY = "somesigningkey"
	SALT       = "salt"
)

func generateHash(password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))

	return fmt.Sprintf("%x", hash.Sum([]byte(SALT)))
}

type tokenClaims struct {
	jwt.RegisteredClaims
	UserID int `json:"id"`
}

type AuthService struct {
	repo repository.Authorization
}

func NewAuthService(repo repository.Authorization) *AuthService {
	return &AuthService{repo: repo}
}

func (as *AuthService) CreateUser(ctx context.Context, user models.User) (int, error) {
	user.Password = generateHash(user.Password)
	return as.repo.CreateUser(ctx, user)
}

func (as *AuthService) GenerateToken(ctx context.Context, login, password string) (string, error) {
	user, err := as.repo.GetUser(ctx, login, generateHash(password))
	if err != nil {
		return ``, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &tokenClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TOKEN_EXP)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: user.ID,
	})

	return token.SignedString([]byte(SECRET_KEY))
}

func (as *AuthService) ParseToken(ctx context.Context, tokenGot string) (int, error) {
	var tokenClaims tokenClaims

	token, err := jwt.ParseWithClaims(tokenGot, &tokenClaims,
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("invalid signing method: %v", t.Header["alg"])
			}
			return []byte(SECRET_KEY), nil
		})

	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, fmt.Errorf("token is not valid")
	}

	return tokenClaims.UserID, nil
}
