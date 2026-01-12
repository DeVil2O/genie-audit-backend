package db

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"

	"genie-audit-backend/logger"
)

var (
	pool    *pgxpool.Pool
	once    sync.Once
	initErr error
)

// GetPool returns a singleton pgx pool, initializing it once.
func GetPool(ctx context.Context, connString string) (*pgxpool.Pool, error) {
	once.Do(func() {
		logger.Log.Info("initializing database pool (singleton)")
		p, err := pgxpool.New(ctx, connString)
		if err != nil {
			initErr = err
			return
		}
		pool = p
	})
	return pool, initErr
}

// Close releases the singleton pool.
func Close() {
	if pool != nil {
		logger.Log.Info("closing database pool")
		pool.Close()
	}
}
