package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"net/http"

	"github.com/sonixxks/Mentor/backend/internal/auth"
	"github.com/sonixxks/Mentor/backend/internal/catalog"
	"github.com/sonixxks/Mentor/backend/internal/chat"
	"github.com/sonixxks/Mentor/backend/internal/config"
	"github.com/sonixxks/Mentor/backend/internal/curator"
	"github.com/sonixxks/Mentor/backend/internal/faq"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres"
	"github.com/sonixxks/Mentor/backend/internal/requests"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/middleware"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/statusroutes"
)

func Run(ctx context.Context, lg *slog.Logger, cfg *config.Config) error {
	lg.Info("Starting Mentor backend...",
		"env", cfg.App.Env,
	)

	//Database
	poolCfg := postgres.PoolConfig{
		MaxConns:   cfg.DB.MaxConns,
		ConnString: cfg.DB.DSN(),
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
	//validator
	//validate := validation.NewValidator()

	//auth
	tokens, err := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.Issuer)
	if err != nil {
		return fmt.Errorf("create token manager: %w", err)
	}
	authMiddleware := middleware.NewAuth(tokens)

	//services
	requestService := requests.NewService(requests.NewPostgresRepository(pool))
	// Модуль подбора (процент совместимости) подключается вторым аргументом, пока nil.
	catalogService := catalog.NewService(catalog.NewPostgresRepository(pool), nil)
	curatorService := curator.NewService(curator.NewPostgresRepository(pool))
	chatService := chat.NewService(chat.NewPostgresRepository(pool))
	faqBot, err := faq.New()
	if err != nil {
		return fmt.Errorf("load faq: %w", err)
	}

	//handlers register
	serviceHandlers := make([]rest.ServiceHandler, 0)
	//serviceHandlers = append(serviceHandlers, auth.NewHandler(validate, lg, &authService))
	serviceHandlers = append(serviceHandlers,
		requests.NewHandler(requestService, authMiddleware, lg),
		catalog.NewHandler(catalogService, authMiddleware, lg),
		curator.NewHandler(curatorService, authMiddleware, lg),
		chat.NewHandler(chatService, authMiddleware, lg),
		faq.NewHandler(faqBot),
	)

	//router
	router := rest.NewRouter(serviceHandlers, "v1")

	//status routes
	healthDB := statusroutes.DBHealthHandler{DBH: pool}
	router.HandleFunc("GET /health", healthDB.Check)

	//Server
	srv := rest.BuildServer(cfg.App.Host, strconv.Itoa(cfg.App.Port), router)

	serverErr := make(chan error, 1)

	go func() {
		lg.Info("Server starting on...",
			"address", srv.Addr,
		)
		if serveErr := srv.ListenAndServe(); serveErr != nil {
			if !errors.Is(serveErr, http.ErrServerClosed) {
				serverErr <- serveErr
			}
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server failed: %w", err)
	case <-ctx.Done():
		lg.Info("shutdown signal received")
	}

	// Отдельный контекст ограничивает время на завершение работы
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	lg.Info("server stopped gracefully")

	return nil
}
