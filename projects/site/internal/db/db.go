package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	_ "modernc.org/sqlite"
)

const driverName = "sqlite"

var ErrEmptyDSN = errors.New("db: empty dsn")

func Init(ctx context.Context, dsn string) (*sql.DB, error) {
	if dsn == "" {
		return nil, ErrEmptyDSN
	}

	conn, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("db: open: %w", err)
	}

	conn.SetMaxOpenConns(1)

	if _, err = conn.ExecContext(ctx, Schema); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("db: apply schema: %w", err)
	}

	return conn, nil
}

func InitMemory(ctx context.Context) (*sql.DB, error) {
	return Init(ctx, ":memory:")
}
