package pgstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Rhymond/go-money"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zasuchilas/gophermart/internal/gophermart/config"
	"github.com/zasuchilas/gophermart/internal/gophermart/logger"
	"github.com/zasuchilas/gophermart/internal/gophermart/models"
	"github.com/zasuchilas/gophermart/internal/gophermart/storage"
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

func (d *PgStorage) Register(ctx context.Context, login, pass string) (userID int64, err error) {
	var id int64
	err = d.db.QueryRowContext(
		ctx,
		"INSERT INTO gophermart.users (login, pass_hash) VALUES($1, $2) ON CONFLICT DO NOTHING RETURNING id",
		login, pass,
	).Scan(&id)
	return id, err
}

func (d *PgStorage) GetLoginData(ctx context.Context, login, password string) (*models.LoginData, error) {
	var v models.LoginData
	err := d.db.QueryRowContext(ctx,
		"SELECT id, login, pass_hash FROM gophermart.users WHERE login = $1 AND deleted = false",
		login).Scan(&v.UserID, &v.Login, &v.PasswordHash)
	return &v, err
}

func (d *PgStorage) RegisterOrder(ctx context.Context, userID int64, orderNum int) error {

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tx, err := d.db.BeginTx(ctxTm, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt1, err := tx.PrepareContext(ctxTm,
		"SELECT user_id FROM gophermart.orders WHERE order_num = $1;")
	if err != nil {
		logger.Log.Error("preparing select stmt", zap.Error(err))
		return err
	}
	defer stmt1.Close()

	stmt2, err := tx.PrepareContext(ctxTm,
		"INSERT INTO gophermart.orders (order_num, user_id) VALUES ($1, $2);")
	if err != nil {
		logger.Log.Error("preparing insert stmt", zap.Error(err))
		return err
	}
	defer stmt2.Close()

	select {
	case <-ctxTm.Done():
		return fmt.Errorf("the operation was canceled on select stmt")
	default:
		var storedUserID int64
		err = stmt1.QueryRowContext(ctxTm, orderNum).Scan(&storedUserID)
		orderFound := err == nil
		errorWithoutNotFound := err != nil && !errors.Is(err, sql.ErrNoRows)
		if orderFound {
			if userID == storedUserID {
				return storage.ErrNumberDone
			}
			return storage.ErrNumberAdded
		} else if errorWithoutNotFound {
			return err
		}
	}

	select {
	case <-ctxTm.Done():
		return fmt.Errorf("the operation was canceled on insert stmt")
	default:
		_, err = stmt2.ExecContext(ctx, orderNum, userID)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (d *PgStorage) GetUserOrders(ctx context.Context, userID int64) (models.OrderData, error) {

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctxTm,
		`SELECT order_num, status, accrual, uploaded_at FROM gophermart.orders WHERE user_id = $1 ORDER BY uploaded_at DESC`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	select {
	case <-ctxTm.Done():
		return nil, fmt.Errorf("the operation was canceled")
	default:
		rows, er := stmt.QueryContext(ctxTm, userID)
		if er != nil {
			return nil, er
		}
		defer rows.Close()

		orders := make(models.OrderData, 0)
		for rows.Next() {
			var (
				v          models.Order
				accrual    int64
				uploadedAt time.Time
			)
			err = rows.Scan(&v.OrderNum, &v.Status, &accrual, &uploadedAt)
			if err != nil {
				return nil, err
			}
			v.Accrual = money.New(accrual, money.RUB).AsMajorUnits()
			v.UploadedAt = uploadedAt.Format(time.RFC3339)
			orders = append(orders, &v)
		}

		err = rows.Err()
		if err != nil {
			return nil, err
		}

		if len(orders) == 0 {
			return nil, storage.ErrNotFound
		}

		return orders, nil
	}
}

func (d *PgStorage) GetUserBalance(ctx context.Context, userID int64) (*models.UserBalance, error) {

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctxTm,
		`SELECT balance, withdrawn FROM gophermart.users WHERE id = $1`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	select {
	case <-ctxTm.Done():
		return nil, fmt.Errorf("the operation was canceled")
	default:
		var (
			v         models.UserBalance
			current   int64
			withdrawn int64
		)
		er := stmt.QueryRowContext(ctxTm, userID).
			Scan(&current, &withdrawn)
		if er != nil {
			return nil, er
		}
		v.Current = money.New(current, money.RUB).AsMajorUnits()
		v.Withdrawn = money.New(withdrawn, money.RUB).AsMajorUnits()
		return &v, nil
	}
}

func (d *PgStorage) WithdrawTransaction(ctx context.Context, userID int64, orderNum int, sum *money.Money) error {
	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tx, err := d.db.BeginTx(ctxTm, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	checkStmt, err := tx.PrepareContext(ctxTm,
		"SELECT balance, withdrawn FROM gophermart.users WHERE id = $1;")
	if err != nil {
		logger.Log.Error("preparing check stmt", zap.Error(err))
		return err
	}
	defer checkStmt.Close()

	withdrawalsStmt, err := tx.PrepareContext(ctxTm,
		"INSERT INTO gophermart.withdrawals (user_id, order_num, amount) VALUES ($1, $2, $3);")
	if err != nil {
		logger.Log.Error("preparing withdrawals stmt", zap.Error(err))
		return err
	}
	defer withdrawalsStmt.Close()

	balanceStmt, err := tx.PrepareContext(ctxTm,
		"UPDATE gophermart.users SET balance = $1, withdrawn = $2 WHERE id = $3;")
	if err != nil {
		logger.Log.Error("preparing balance stmt", zap.Error(err))
		return err
	}
	defer balanceStmt.Close()

	var (
		balance   int64
		withdrawn int64
	)

	select {
	case <-ctxTm.Done():
		return fmt.Errorf("the operation was canceled on check stmt")
	default:
		err = checkStmt.QueryRowContext(ctxTm, userID).Scan(&balance, &withdrawn)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return fmt.Errorf("user not found (userID %d", userID)
			}
			return err
		}
		currentBalance := money.New(balance, money.RUB)
		nextBalance, er := currentBalance.Subtract(sum)
		if er != nil {
			return fmt.Errorf("error in calculating the new balance (current balance %f, sum %f)",
				currentBalance.AsMajorUnits(), sum.AsMajorUnits())
		}
		if nextBalance.IsNegative() {
			return storage.ErrNotEnoughFunds
		}
		currentWithdrawn := money.New(withdrawn, money.RUB)
		nextWithdrawn, er := currentWithdrawn.Add(sum)
		if er != nil {
			return fmt.Errorf("error in calculating the new withdrawn (current withdrawn %f, sum %f)",
				currentWithdrawn.AsMajorUnits(), sum.AsMajorUnits())
		}

		// new values for db
		balance = nextBalance.Amount()
		withdrawn = nextWithdrawn.Amount()
	}

	select {
	case <-ctxTm.Done():
		return fmt.Errorf("the operation was canceled on withdrawals stmt")
	default:
		_, err = withdrawalsStmt.ExecContext(ctx,
			userID, orderNum, sum.Amount(),
		)
		if err != nil {
			return err
		}
	}

	select {
	case <-ctxTm.Done():
		return fmt.Errorf("the operation was canceled on balance stmt")
	default:
		_, err = balanceStmt.ExecContext(ctx,
			balance, withdrawn, userID,
		)
		if err != nil {
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}

func (d *PgStorage) GetUserWithdrawals(ctx context.Context, userID int64) (models.WithdrawalsData, error) {

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctxTm,
		`SELECT order_num, amount, processed_at FROM gophermart.withdrawals WHERE user_id = $1 ORDER BY processed_at DESC`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	select {
	case <-ctxTm.Done():
		return nil, fmt.Errorf("the operation was canceled")
	default:
		rows, er := stmt.QueryContext(ctxTm, userID)
		if er != nil {
			return nil, er
		}
		defer rows.Close()

		withdrawals := make(models.WithdrawalsData, 0)
		for rows.Next() {
			var (
				v           models.Withdrawal
				amount      int64
				processedAt time.Time
			)
			err = rows.Scan(&v.OrderNum, &amount, &processedAt)
			if err != nil {
				return nil, err
			}
			v.Sum = money.New(amount, money.RUB).AsMajorUnits()
			v.ProcessedAt = processedAt.Format(time.RFC3339)
			withdrawals = append(withdrawals, &v)
		}

		err = rows.Err()
		if err != nil {
			return nil, err
		}

		if len(withdrawals) == 0 {
			return nil, storage.ErrNotFound
		}

		return withdrawals, nil
	}
}

func (d *PgStorage) GetOrdersPack(ctx context.Context) ([]*models.OrderRow, error) {
	statuses := []string{
		storage.OrderStatusNew,
		storage.OrderStatusRegistered,
		storage.OrderStatusProcessing,
	}
	limit := config.WorkerPackLimit // TODO: get limit from above

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	stmt, err := d.db.PrepareContext(ctxTm,
		`SELECT id, order_num, status, accrual, user_id, uploaded_at FROM gophermart.orders WHERE status = any($1) LIMIT $2`)
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

		orders := make([]*models.OrderRow, 0)
		for rows.Next() {
			var (
				gd models.OrderRow
			)
			err = rows.Scan(&gd.ID, &gd.OrderNum, &gd.Status, &gd.Accrual, &gd.UserID, &gd.UploadedAt)
			if err != nil {
				return nil, err
			}
			orders = append(orders, &gd)
		}

		err = rows.Err()
		if err != nil {
			return nil, err
		}

		return orders, nil
	}
}

func (d *PgStorage) UpdateOrder(ctx context.Context, userID, id int64, status string, accrual float64) error {
	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tx, err := d.db.BeginTx(ctxTm, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	orderStmt, err := tx.PrepareContext(ctxTm,
		"UPDATE gophermart.orders SET status = $1, accrual = $2 WHERE id = $3;")
	if err != nil {
		logger.Log.Info("preparing order stmt", zap.String("error", err.Error()))
		return err
	}
	defer orderStmt.Close()

	balanceStmt, err := tx.PrepareContext(ctxTm,
		"UPDATE gophermart.users SET balance = balance + $1 WHERE id = $2;")
	if err != nil {
		logger.Log.Info("preparing balance stmt", zap.String("error", err.Error()))
		return err
	}
	defer balanceStmt.Close()

	select {
	case <-ctxTm.Done():
		return fmt.Errorf("the operation was canceled")
	default:
		_, err = orderStmt.ExecContext(ctx,
			status, accrual, id,
		)
		if err != nil {
			return err
		}
	}

	if status == storage.OrderStatusProcessed {
		select {
		case <-ctxTm.Done():
			return fmt.Errorf("the operation was canceled")
		default:
			_, err = balanceStmt.ExecContext(ctx,
				accrual, userID,
			)
			if err != nil {
				return err
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
