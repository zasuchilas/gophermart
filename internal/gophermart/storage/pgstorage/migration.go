package pgstorage

import (
	"context"
	"database/sql"
	"github.com/zasuchilas/gophermart/internal/gophermart/logger"
	"go.uber.org/zap"
	"time"
)

// TODO: goose

func createTablesIfNeed(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	q := `
		CREATE SCHEMA IF NOT EXISTS gophermart;

		CREATE TABLE IF NOT EXISTS gophermart.users (
			id SERIAL PRIMARY KEY,
			login VARCHAR(254) NOT NULL UNIQUE,
			pass_hash VARCHAR(254) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
			deleted BOOL NOT NULL DEFAULT false,
		  balance INTEGER NOT NULL DEFAULT 0,
		  withdrawn INTEGER NOT NULL DEFAULT 0
		);
		CREATE INDEX IF NOT EXISTS idx_login ON gophermart.users (login);
		CREATE INDEX IF NOT EXISTS idx_deleted ON gophermart.users (deleted);

		CREATE TABLE IF NOT EXISTS gophermart.orders (
			id SERIAL PRIMARY KEY,
			order_num INT8 NOT NULL UNIQUE,
			status VARCHAR(25) NOT NULL DEFAULT 'NEW',
			accrual INTEGER NOT NULL DEFAULT 0,
		  user_id INT8 REFERENCES users (id),
			uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
		);
		CREATE INDEX IF NOT EXISTS idx_status ON gophermart.orders (status);

		CREATE TABLE IF NOT EXISTS gophermart.withdrawals (
		  id SERIAL PRIMARY KEY,
		  user_id INT8 REFERENCES users (id),
		  order_num INT8 NOT NULL, -- not related to table orders
		  amount INT NOT NULL DEFAULT 0,
		  processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
		);
		
  `

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		logger.Log.Fatal("creating postgresql tables", zap.String("error", err.Error()))
	}
}
