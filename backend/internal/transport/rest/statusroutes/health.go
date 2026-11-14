package statusroutes

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

type DBHealthHandler struct {
	DBH Pinger
}

func (healthHandler *DBHealthHandler) Check(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()
	err := healthHandler.DBH.Ping(ctx)
	if err != nil {
		slog.Error("database ping failed", slog.Any("error", err))
		http.Error(w, "database unavailable", http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_, writeErr := w.Write([]byte(`{"status":"ok"}`))
	if writeErr != nil {
		slog.Error("health check: cannot write response",
			slog.Any("error", writeErr),
			slog.String("remote_addr", r.RemoteAddr),
		)
		return
	}
}

type Pinger interface {
	Ping(context.Context) error
}
