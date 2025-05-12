package database

import (
	"context"
	"database/sql"
	"errors"

	"github.com/bayu-gara/parking-lot/pkg/config"
	sqlite "github.com/bayu-gara/parking-lot/pkg/database/sqlite"
)

type SQLDB interface {
	Begin() (*sql.Tx, error)
	BeginTx(ctx context.Context, opts *sql.TxOptions) (*sql.Tx, error)
	Close() error
	Conn(ctx context.Context) (*sql.Conn, error)
	Exec(query string, args ...any) (sql.Result, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	Ping() error
	PingContext(ctx context.Context) error
	Prepare(query string) (*sql.Stmt, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func Init(cfg config.DatabaseConfig) (SQLDB, error) {
	switch cfg.Engine {
	case "sqlite":
		return sqlite.Init(cfg.SQLiteConfig)
	}

	return nil, errors.New("Unsupported database engine: " + cfg.Engine)
}
