package pgstorage

import (
	"context"
	"database/sql"
	"github.com/Rhymond/go-money"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zasuchilas/gophermart/internal/accrual/config"
	"github.com/zasuchilas/gophermart/internal/accrual/logger"
	"github.com/zasuchilas/gophermart/internal/accrual/models"
	"github.com/zasuchilas/gophermart/internal/accrual/storage"
	"go.uber.org/zap"
)

type PgStorage struct {
	db *sql.DB
}

func New() *PgStorage {
	if config.DatabaseURI == "" {
		logger.Log.Fatal("database connection string is empty")
	}

	db, err := sql.Open("pgx", config.DatabaseURI)
	if err != nil {
		logger.Log.Fatal("opening connection to postgresql", zap.Error(err))
		return nil
	}

	logger.Log.Debug("creating db tables if need")
	createTablesIfNeed(db) // TODO: goose

	return &PgStorage{
		db: db,
	}
}

func (d *PgStorage) Stop() {
	if d.db != nil {
		_ = d.db.Close()
	}
}

func (d *PgStorage) InstanceName() string {
	return storage.InstancePostgresql
}

func (d *PgStorage) RegisterNewGoods(ctx context.Context, match, rewardType string, reward int) (int64, error) {
	var id int64
	err := d.db.QueryRowContext(
		ctx,
		"INSERT INTO accrual.goods (match, reward, reward_type) VALUES($1, $2, $3) ON CONFLICT DO NOTHING RETURNING id",
		match, reward, rewardType,
	).Scan(&id)
	return id, err
}

func (d *PgStorage) RegisterNewOrder(ctx context.Context, orderNum int, receipt string) (int64, error) {
	var id int64
	err := d.db.QueryRowContext(
		ctx,
		"INSERT INTO accrual.orders (order_num, receipt) VALUES($1, $2) ON CONFLICT DO NOTHING RETURNING id",
		orderNum, receipt,
	).Scan(&id)
	return id, err
}

func (d *PgStorage) GetOrderData(ctx context.Context, orderNum int) (*models.OrderData, error) {
	var (
		v       models.OrderData
		accrual int64
	)
	err := d.db.QueryRowContext(ctx,
		"SELECT order_num, status, accrual FROM gophermart.orders WHERE order_num = $1",
		orderNum).Scan(&v.Order, &v.Status, &accrual)

	v.Accrual = money.New(accrual, money.RUB).AsMajorUnits()

	return &v, err
}
