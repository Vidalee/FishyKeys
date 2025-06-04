package db

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
)

var globalPool *pgxpool.Pool

func Init(ctx context.Context, dsn string) error {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("create pgx pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping pgx pool: %w", err)
	}
	globalPool = pool
	return nil
}

func Pool() *pgxpool.Pool {
	if globalPool == nil {
		panic("pgx pool is not initialized")
	}
	return globalPool
}
