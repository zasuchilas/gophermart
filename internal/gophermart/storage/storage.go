package storage

import (
	"context"
	"errors"
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
}
