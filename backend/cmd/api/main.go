package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/sonixxks/Mentor/backend/internal/app"
	"github.com/sonixxks/Mentor/backend/internal/config"
	"github.com/sonixxks/Mentor/backend/internal/platform/logger"
)

func main() {
	appError := run()
	if appError != nil {
		slog.Error("application failed", "error", appError)
		os.Exit(1)
	}
}

func run() error {
	if err := godotenv.Load(); errors.Is(err, fs.ErrNotExist) {
		log.Println(".env не найден, используются переменные окружения")
	} else if err != nil {
		return fmt.Errorf("load .env: %w", err)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	lg, slogErr := logger.SetupLogger(cfg.App.Env)
	if slogErr != nil {
		return fmt.Errorf("setup logger: %w", slogErr)
	}
	slog.SetDefault(lg)

	ctx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,  // Ctrl+C
		syscall.SIGTERM, // docker stop / SIGTERM от оркестратора
	)
	defer stop()

	runError := app.Run(ctx, lg, cfg)
	if runError != nil {
		return runError
	}
	return nil
}
