package handler

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"

	"global-ranks/internal/config"
	"global-ranks/internal/middleware"
)

// Deps holds shared dependencies injected into all handlers.
type Deps struct {
	DB             *sql.DB
	Cfg            config.Config
	ScoreRateLimit *middleware.RateLimiter
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to write json response", "error", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
