package pgstorage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Rhymond/go-money"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zasuchilas/gophermart/internal/accrual/config"
	"github.com/zasuchilas/gophermart/internal/accrual/logger"
	"github.com/zasuchilas/gophermart/internal/accrual/models"
	"github.com/zasuchilas/gophermart/internal/accrual/storage"
	"github.com/zasuchilas/gophermart/internal/common"
	"go.uber.org/zap"
	"time"
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

func (d *PgStorage) RegisterNewGoods(ctx context.Context, match, rewardType string, reward float64) (int64, error) {
	var id int64
	err := d.db.QueryRowContext(
		ctx,
		"INSERT INTO accrual.goods (match, reward, reward_type) VALUES($1, $2, $3) ON CONFLICT DO NOTHING RETURNING id",
		match, reward, rewardType,
	).Scan(&id)
	return id, err
}

func (d *PgStorage) RegisterNewOrder(ctx context.Context, orderNum string, receipt string) (int64, error) {
	var id int64
	err := d.db.QueryRowContext(
		ctx,
		"INSERT INTO accrual.orders (order_num, receipt) VALUES($1, $2) ON CONFLICT DO NOTHING RETURNING id",
		orderNum, receipt,
	).Scan(&id)
	return id, err
}

func (d *PgStorage) GetOrderData(ctx context.Context, orderNum string) (*models.OrderData, error) {
	var (
		v       models.OrderData
		accrual int64
	)
	err := d.db.QueryRowContext(ctx,
		"SELECT order_num, status, accrual FROM accrual.orders WHERE order_num = $1",
		orderNum).Scan(&v.Order, &v.Status, &accrual)

	v.Accrual = money.New(accrual, money.RUB).AsMajorUnits()

	return &v, err
}

func (d *PgStorage) GetGoods(ctx context.Context) ([]*models.GoodsData, error) {

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctxTm, `SELECT match, reward, reward_type FROM accrual.goods WHERE deleted = false`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	select {
	case <-ctxTm.Done():
		return nil, fmt.Errorf("the operation was canceled")
	default:
		rows, er := stmt.QueryContext(ctxTm)
		if er != nil {
			return nil, er
		}
		defer rows.Close()

		goods := make([]*models.GoodsData, 0)
		for rows.Next() {
			var gd models.GoodsData
			err = rows.Scan(&gd.Match, &gd.Reward, &gd.RewardType)
			if err != nil {
				return nil, err
			}
			goods = append(goods, &gd)
		}

		err = rows.Err()
		if err != nil {
			return nil, err
		}

		return goods, nil
	}
}

func (d *PgStorage) GetOrders(ctx context.Context) ([]*models.AccrualOrder, error) {
	statuses := []string{
		common.OrderStatusNew,
		common.OrderStatusRegistered,
		common.OrderStatusProcessing,
	}
	limit := config.WorkerPackLimit

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctxTm,
		`SELECT id, order_num, status, accrual, receipt, uploaded_at FROM accrual.orders WHERE status = any($1) LIMIT $2`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	select {
	case <-ctxTm.Done():
		return nil, fmt.Errorf("the operation was canceled")
	default:
		rows, er := stmt.QueryContext(ctxTm, statuses, limit)
		if er != nil {
			return nil, er
		}
		defer rows.Close()

		orders := make([]*models.AccrualOrder, 0)
		for rows.Next() {
			var (
				ord     models.AccrualOrder
				accrual int64
				receipt string
				rc      models.Receipt
			)
			err = rows.Scan(&ord.ID, &ord.OrderNum, &ord.Status, &accrual, &receipt, &ord.UploadedAt)
			if err != nil {
				return nil, err
			}
			err = json.Unmarshal([]byte(receipt), &rc)
			if err == nil {
				ord.Receipt = &rc
			}
			ord.Accrual = money.New(accrual, money.RUB)
			orders = append(orders, &ord)
		}

		err = rows.Err()
		if err != nil {
			return nil, err
		}

		return orders, nil
	}
}

func (d *PgStorage) UpdateOrder(ctx context.Context, id int64, status string, accrual *money.Money) error {
	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctxTm,
		"UPDATE accrual.orders SET status = $1, accrual = $2 WHERE id = $3;")
	if err != nil {
		logger.Log.Error("preparing balance stmt", zap.Error(err))
		return err
	}
	defer stmt.Close()

	select {
	case <-ctxTm.Done():
		return fmt.Errorf("the operation was canceled")
	default:
		_, err = stmt.ExecContext(ctx,
			status, accrual.Amount(), id,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
