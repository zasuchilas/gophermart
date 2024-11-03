package pgstorage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zasuchilas/gophermart/internal/gophermart/config"
	"github.com/zasuchilas/gophermart/internal/gophermart/logger"
	"github.com/zasuchilas/gophermart/internal/gophermart/model"
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
		"INSERT INTO users (login, pass_hash) VALUES($1, $2) ON CONFLICT DO NOTHING RETURNING id",
		login, pass,
	).Scan(&id)
	return id, err
}

func (d *PgStorage) GetLoginData(ctx context.Context, login, password string) (*model.LoginData, error) {
	var v model.LoginData
	err := d.db.QueryRowContext(ctx,
		"SELECT id, login, pass_hash FROM users WHERE login = $1 AND deleted = false",
		login).Scan(&v.UserID, &v.Login, &v.PasswordHash)
	return &v, err
}

func (d *PgStorage) RegisterOrder(ctx context.Context, userID int64, number int) error {

	ctxTm, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	tx, err := d.db.BeginTx(ctxTm, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt1, err := tx.PrepareContext(ctxTm,
		"SELECT user_id FROM orders WHERE number = $1;")
	if err != nil {
		logger.Log.Error("preparing select stmt", zap.Error(err))
		return err
	}
	defer stmt1.Close()

	stmt2, err := tx.PrepareContext(ctxTm,
		"INSERT INTO orders (number, user_id) VALUES ($1, $2);")
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
		err = stmt1.QueryRowContext(ctxTm, number).Scan(&storedUserID)
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
		_, err = stmt2.ExecContext(ctx, number, userID)
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
