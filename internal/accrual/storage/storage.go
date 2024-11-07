package storage

import (
	"context"
	"github.com/Rhymond/go-money"
	"github.com/zasuchilas/gophermart/internal/accrual/models"
)

const (
	InstancePostgresql = "pgsql"
)

type Storage interface {
	Stop()
	InstanceName() string

	RegisterNewGoods(ctx context.Context, match, rewardType string, reward float64) (int64, error)
	RegisterNewOrder(ctx context.Context, orderNum string, receipt string) (int64, error)
	GetOrderData(ctx context.Context, orderNum string) (*models.OrderData, error)

	GetGoods(ctx context.Context) ([]*models.GoodsData, error)
	GetOrders(ctx context.Context) ([]*models.AccrualOrder, error)
	UpdateOrder(ctx context.Context, id int64, status string, accrual *money.Money) error
}
