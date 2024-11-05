package storage

import (
	"context"
	"github.com/zasuchilas/gophermart/internal/accrual/models"
)

const (
	InstancePostgresql = "pgsql"

	OrderStatusNew        = "NEW"        // gophermart service
	OrderStatusRegistered = "REGISTERED" // accrual service
	OrderStatusProcessing = "PROCESSING"
	OrderStatusInvalid    = "INVALID"
	OrderStatusProcessed  = "PROCESSED"
)

type Storage interface {
	Stop()
	InstanceName() string

	RegisterNewGoods(ctx context.Context, match, rewardType string, reward float64) (int64, error)
	RegisterNewOrder(ctx context.Context, orderNum int, receipt string) (int64, error)
	GetOrderData(ctx context.Context, orderNum int) (*models.OrderData, error)

	GetGoods(ctx context.Context) ([]*models.GoodsData, error)
	GetOrders(ctx context.Context) ([]*models.AccrualOrder, error)
	UpdateOrder(ctx context.Context, id int64, status string, accrual float64) error
}
