package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"net/http"

	"github.com/sonixxks/Mentor/backend/internal/account"
	"github.com/sonixxks/Mentor/backend/internal/auth"
	"github.com/sonixxks/Mentor/backend/internal/catalog"
	"github.com/sonixxks/Mentor/backend/internal/chat"
	"github.com/sonixxks/Mentor/backend/internal/config"
	"github.com/sonixxks/Mentor/backend/internal/curator"
	"github.com/sonixxks/Mentor/backend/internal/platform/mailer"
	"github.com/sonixxks/Mentor/backend/internal/platform/memory"
	"github.com/sonixxks/Mentor/backend/internal/platform/postgres"
	"github.com/sonixxks/Mentor/backend/internal/platform/storage"
	"github.com/sonixxks/Mentor/backend/internal/platform/validation"
	"github.com/sonixxks/Mentor/backend/internal/profile"
	"github.com/sonixxks/Mentor/backend/internal/profile/profilehttp"
	"github.com/sonixxks/Mentor/backend/internal/requests"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/middleware"
	"github.com/sonixxks/Mentor/backend/internal/transport/rest/statusroutes"
	"github.com/sonixxks/Mentor/backend/internal/user"
	"github.com/sonixxks/Mentor/backend/internal/workspaces"
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

	//Settings
	codeSettings := auth.CodeSettings{
		TTL:            cfg.ConfirmationCode.ExpiresInSeconds,
		ResendInterval: cfg.ConfirmationCode.ResendAfterSeconds,
		MaxAttempts:    cfg.ConfirmationCode.MaxAttempts,
		HashSecret:     []byte(cfg.ConfirmationCode.OTPSecret),
	}
	refreshTokenSettings := auth.RefreshTokenSettings{
		RefreshTTL:         cfg.RefreshTokens.RefreshTTL,
		RefreshRememberTTL: cfg.RefreshTokens.RefreshRememberTTL,
	}

	//TransactionManager
	transactionManager := postgres.NewTransactionManager(pool)

	//validator
	validate := validation.NewValidator()

	//repositories
	userRepository := user.NewRepository(transactionManager)
	profileRepository := profile.NewRepository(transactionManager)

	//TokenManager
	tokenManager, err := auth.NewTokenManager(cfg.JWT.Secret, cfg.JWT.AccessTTL, cfg.JWT.Issuer)
	if err != nil {
		return fmt.Errorf("create token manager: %w", err)
	}
	authMiddleware := middleware.NewAuth(tokenManager)
	//Hasher
	hasher, err := auth.NewDefaultPasswordHasher()
	if err != nil {
		return fmt.Errorf("create password hasher: %w", err)
	}

	devMailer := mailer.NewLogMailer(lg)

	//uploads
	files, err := storage.NewLocal(cfg.Uploads.Dir, cfg.Uploads.BaseURL)
	if err != nil {
		return fmt.Errorf("create uploads storage: %w", err)
	}
	profileService := profile.NewService(profileRepository, files)

	//services

	authService := auth.NewService(
		auth.ServiceDeps{
			Users:         userRepository,
			Profiles:      profileRepository,
			Tokens:        tokenManager,
			RefreshTokens: memory.NewRefreshTokenStore(),
			Hasher:        hasher,
			Transactions:  transactionManager,
			Registrations: memory.NewRegistrationStore(),
			ResendLimiter: memory.NewResendLimiter(),
			Mailer:        devMailer,
			Codes:         auth.RandomCodeGenerator{},

			PasswordResets: memory.NewPasswordResetStore(),
			ResetMailer:    devMailer,
		},
		auth.Settings{CodeSettings: codeSettings, RefreshTokenSettings: refreshTokenSettings},
	)
	requestService := requests.NewService(requests.NewPostgresRepository(pool))
	// Модуль подбора (процент совместимости) подключается вторым аргументом, пока nil.
	catalogService := catalog.NewService(catalog.NewPostgresRepository(pool), nil)
	curatorService := curator.NewService(curator.NewPostgresRepository(pool))
	chatService := chat.NewService(chat.NewPostgresRepository(pool))

	//handlers register
	serviceHandlers := make([]rest.ServiceHandler, 0)
	serviceHandlers = append(serviceHandlers,
		auth.NewHandler(validate, lg, authService),
		account.NewHandler(userRepository, authMiddleware, lg),
		profilehttp.NewHandler(profileService, authMiddleware, lg),
	)
	serviceHandlers = append(serviceHandlers,
		requests.NewHandler(requestService, authMiddleware, lg),
		catalog.NewHandler(catalogService, authMiddleware, lg),
		curator.NewHandler(curatorService, authMiddleware, lg),
		chat.NewHandler(chatService, authMiddleware, lg),
		workspaces.NewHandler(workspaces.NewRepository(pool), authMiddleware, lg),
	)

	//router
	router := rest.NewRouter(serviceHandlers, "v1")

	//status routes
	healthDB := statusroutes.DBHealthHandler{DBH: pool}
	router.HandleFunc("GET /health", healthDB.Check)
	router.Handle("GET /uploads/", files.Handler("/uploads/"))

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
