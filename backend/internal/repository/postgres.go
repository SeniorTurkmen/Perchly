package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	pgxvec "github.com/pgvector/pgvector-go/pgx"
)

// NewPostgresPool builds a pgx connection pool for the given DSN.
//
// Pool creation itself does not require a live connection (pgx connects
// lazily on first use), so this only fails on malformed configuration.
// Actual reachability should be checked separately, e.g. via Ping in a
// health check.
func NewPostgresPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse postgres dsn: %w", err)
	}

	// Registers the pgvector `vector` codec on every pooled connection, so
	// []float32 <-> vector(N) works via pgvector.NewVector without manual
	// text encoding.
	poolCfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		return pgxvec.RegisterTypes(ctx, conn)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	return pool, nil
}
