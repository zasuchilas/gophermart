package pgstorage

import (
	"context"
	"database/sql"
	"github.com/zasuchilas/gophermart/internal/gophermart/logger"
	"time"
)

// TODO: goose

func createTablesIfNeed(db *sql.DB) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	q := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			login VARCHAR(254) NOT NULL UNIQUE,
			pass_hash VARCHAR(254) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT now(),
			deleted BOOL NOT NULL DEFAULT false
		);
		CREATE INDEX IF NOT EXISTS idx_login ON users (login);
		CREATE INDEX IF NOT EXISTS idx_deleted ON users (deleted);

		CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			number INT8 NOT NULL UNIQUE,
			status VARCHAR(25) NOT NULL DEFAULT 'NEW',
			accrual INTEGER NOT NULL DEFAULT 0,
		  user_id INT8 REFERENCES users (id),
			uploaded_at TIMESTAMP NOT NULL DEFAULT now()
		);
		CREATE INDEX IF NOT EXISTS idx_status ON orders (status);
		
  `

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		logger.Log.Fatal("creating postgresql tables")
	}
}
