package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func CreatePool(ctx context.Context, cfg PoolConfig) (*pgxpool.Pool, error) {
	parsedConfig, err := pgxpool.ParseConfig(cfg.ConnString)
	if err != nil {
		return nil, fmt.Errorf("pgxpool cannot ParseConfig, %w", err)
	}

	parsedConfig.MaxConns = cfg.MaxConns
	// parsedConfig.MaxConnLiftime=
	// parsedConfig.MaxConnIdleTime=
	// parsedConfig.HealthCheckPeriod=
	// parsedConfig.ConnConfig.ConnectTimeout=

	pool, err := pgxpool.NewWithConfig(ctx, parsedConfig)
	if err != nil {
		return nil, fmt.Errorf("pgxpool cannot init, %w", err)
	}

	timeoutCtx, cancelFunc := context.WithTimeout(ctx, 5*time.Second)
	defer cancelFunc()
	pingError := pool.Ping(timeoutCtx)
	if pingError != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database error, %w", pingError)
	}

	return pool, nil
}
