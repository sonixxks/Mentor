package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/sonixxks/Mentor/backend/internal/config"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres"
	httptransport "github.com/sonixxks/Mentor/backend/internal/transport/http"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
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

	router := httptransport.NewRouter(httptransport.Deps{
		Logger: lg,
		DB:     pool,
	})

	srv := &http.Server{
		Addr:              net.JoinHostPort("", strconv.Itoa(cfg.App.Port)),
		Handler:           router,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverErr := make(chan error, 1)
	go func() {
		lg.Info("HTTP server started", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("http server: %w", err)
	case <-ctx.Done():
		lg.Info("Shutting down HTTP server...")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("http server shutdown: %w", err)
	}

	lg.Info("HTTP server stopped")
	return nil
}
