package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sonixxks/Mentor/backend/internal/config"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres"
)

func Run(ctx context.Context, lg *slog.Logger, cfg *config.Config) error {
	lg.Info("Starting Mentor backend...",
		"env", cfg.App.Env,
	)

	poolCfg := postgres.PoolConfig{
		MaxConns:   cfg.Db.MaxConns,
		ConnString: cfg.Db.DSN(),
	}

	pool, err := postgres.CreatePool(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("create postgres pool: %w", err)
	}
	defer pool.Close()

	lg.Info("Database connected",
		"host", pool.Config().ConnConfig.Host,
		"port", pool.Config().ConnConfig.Port,
	)
	return nil
}
