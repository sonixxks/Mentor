package http

import (
	"context"
	"net/http"
	"time"
)

const healthDBTimeout = 2 * time.Second

func healthHandler(db Pinger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), healthDBTimeout)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			writeError(w, http.StatusServiceUnavailable, "db_unavailable", "database is unavailable")
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}
