package logger

import (
	"fmt"
	"log/slog"
	"os"
)

func SetupLogger(env string) (*slog.Logger, error) {
	var log *slog.Logger

	switch env {
	case "development":
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	case "production":
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	default:
		return nil, fmt.Errorf("cannot setup Logger, check APP_ENV - development/production")
	}

	return log, nil
}
