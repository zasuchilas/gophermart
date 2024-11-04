package storage

import (
	"context"
	"github.com/zasuchilas/gophermart/internal/accrual/models"
)

const (
	InstancePostgresql = "pgsql"
)

type Storage interface {
	Stop()
	InstanceName() string

	RegisterNewGoods(ctx context.Context, match, rewardType string, reward int) (int64, error)
	RegisterNewOrder(ctx context.Context, orderNum int, receipt string) (int64, error)
	GetOrderData(ctx context.Context, orderNum int) (*models.OrderData, error)
}
