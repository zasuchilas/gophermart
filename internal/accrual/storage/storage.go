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

	ErrNumberDone     = errors.New("number already done by current user")
	ErrNumberAdded    = errors.New("number already added by another user")
	ErrNotEnoughFunds = errors.New("not enough funds on the balance")
)

type Storage interface {
	Stop()
	InstanceName() string

	RegisterNewGoods(ctx context.Context, match, rewardType string, reward int) (int64, error)
}
