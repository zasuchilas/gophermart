package pgstorage

import (
	"context"
	"database/sql"
	"github.com/zasuchilas/gophermart/internal/accrual/logger"
	"time"
)

// TODO: goose

func createTablesIfNeed(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	q := `
		CREATE SCHEMA IF NOT EXISTS accrual;

		CREATE TABLE IF NOT EXISTS accrual.orders (
			id SERIAL PRIMARY KEY,
			order_num INT8 NOT NULL UNIQUE,
			status VARCHAR(25) NOT NULL DEFAULT 'NEW',
			accrual INTEGER NOT NULL DEFAULT 0,
		  receipt TEXT NOT NULL DEFAULT '',
			uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
		);
		CREATE INDEX IF NOT EXISTS idx_status ON accrual.orders (status);

		CREATE TABLE IF NOT EXISTS accrual.goods (
		  id SERIAL PRIMARY KEY,
		  match VARCHAR(254) NOT NULL,
		  reward INT NOT NULL,
		  reward_type VARCHAR(25) NOT NULL,
		  created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
		);
		
  `

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		logger.Log.Fatal("creating postgresql tables")
	}
}
