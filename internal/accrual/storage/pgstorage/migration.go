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
-- 		CREATE TABLE IF NOT EXISTS users (
-- 			id SERIAL PRIMARY KEY,
-- 			login VARCHAR(254) NOT NULL UNIQUE,
-- 			pass_hash VARCHAR(254) NOT NULL,
-- 			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(),
-- 			deleted BOOL NOT NULL DEFAULT false,
-- 		  balance INTEGER NOT NULL DEFAULT 0,
-- 		  withdrawn INTEGER NOT NULL DEFAULT 0
-- 		);
-- 		CREATE INDEX IF NOT EXISTS idx_login ON users (login);
-- 		CREATE INDEX IF NOT EXISTS idx_deleted ON users (deleted);
-- 
-- 		CREATE TABLE IF NOT EXISTS orders (
-- 			id SERIAL PRIMARY KEY,
-- 			order_num INT8 NOT NULL UNIQUE,
-- 			status VARCHAR(25) NOT NULL DEFAULT 'NEW',
-- 			accrual INTEGER NOT NULL DEFAULT 0,
-- 		  user_id INT8 REFERENCES users (id),
-- 			uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
-- 		);
-- 		CREATE INDEX IF NOT EXISTS idx_status ON orders (status);
-- 
-- 		CREATE TABLE IF NOT EXISTS withdrawals (
-- 		  id SERIAL PRIMARY KEY,
-- 		  user_id INT8 REFERENCES users (id),
-- 		  order_num INT8 NOT NULL, -- not related to table orders
-- 		  amount INT NOT NULL DEFAULT 0,
-- 		  processed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
-- 		);
		
  `

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		logger.Log.Fatal("creating postgresql tables")
	}
}
