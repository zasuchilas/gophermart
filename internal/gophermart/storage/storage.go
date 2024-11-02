package storage

import (
	"context"
	"errors"
	"github.com/zasuchilas/gophermart/internal/gophermart/model"
)

const (
	InstancePostgresql = "pgsql"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrGone       = errors.New("deleted")
	ErrBadRequest = errors.New("bad request")
)

type Storage interface {
	Stop()
	InstanceName() string

	Register(ctx context.Context, login, passHash string) (int64, error)
	GetLoginData(ctx context.Context, login, password string) (*model.LoginData, error)
}
