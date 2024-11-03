package storage

import (
	"context"
	"errors"
	"github.com/zasuchilas/gophermart/internal/gophermart/models"
)

const (
	InstancePostgresql = "pgsql"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrGone       = errors.New("deleted")
	ErrBadRequest = errors.New("bad request")

	ErrNumberDone  = errors.New("number already done by current user")
	ErrNumberAdded = errors.New("number already added by another user")
)

type Storage interface {
	Stop()
	InstanceName() string

	Register(ctx context.Context, login, passHash string) (int64, error)
	GetLoginData(ctx context.Context, login, password string) (*models.LoginData, error)
	RegisterOrder(ctx context.Context, userID int64, number int) error
	GetUserOrders(ctx context.Context, userID int64) (models.OrderData, error)
	GetUserBalance(ctx context.Context, userID int64) (*models.UserBalance, error)
}
