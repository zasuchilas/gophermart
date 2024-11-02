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
			deleted BOOL NOT NULL DEFAULT false
		);
		CREATE INDEX IF NOT EXISTS idx_name ON users (login);
		CREATE INDEX IF NOT EXISTS idx_deleted ON users (deleted);
  `

	_, err := db.ExecContext(ctx, q)
	if err != nil {
		logger.Log.Fatal("creating postgresql tables")
	}
}
