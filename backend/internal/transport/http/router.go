package http

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const requestTimeout = 30 * time.Second

// Pinger — всё, что умеет проверять доступность (например, *pgxpool.Pool).
type Pinger interface {
	Ping(ctx context.Context) error
}

type Deps struct {
	Logger *slog.Logger
	DB     Pinger
}

func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(requestLogger(deps.Logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(requestTimeout))

	r.NotFound(func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusNotFound, "not_found", "route not found")
	})
	r.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
		writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	})

	r.Get("/health", healthHandler(deps.DB))

	r.Route("/api/v1", func(_ chi.Router) {
		// сюда монтируются роуты модулей: auth, profiles, requests, ...
	})

	return r
}
