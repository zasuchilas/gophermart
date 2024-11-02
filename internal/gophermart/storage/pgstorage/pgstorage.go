package pgstorage

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/zasuchilas/gophermart/internal/gophermart/config"
	"github.com/zasuchilas/gophermart/internal/gophermart/logger"
	"github.com/zasuchilas/gophermart/internal/gophermart/model"
	"github.com/zasuchilas/gophermart/internal/gophermart/storage"
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
