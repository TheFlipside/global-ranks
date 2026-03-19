package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"global-ranks/internal/config"
	"global-ranks/internal/database"
	"global-ranks/internal/handler"
	"global-ranks/internal/middleware"
	"global-ranks/migrations"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg := config.Load()

	db, err := database.Connect(cfg.DBDSN)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(db, migrations.FS); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	scoreRL := middleware.NewRateLimiter(cfg.RateScorePerSec, cfg.RateScoreBurst)
	generalRL := middleware.NewRateLimiter(cfg.RateGeneralPerSec, cfg.RateGeneralBurst)

	deps := &handler.Deps{
		DB:             db,
		Cfg:            cfg,
		ScoreRateLimit: scoreRL,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := db.PingContext(ctx); err != nil {
			http.Error(w, `{"status":"error","db":"unreachable"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","db":"ok"}`))
	})

	mux.HandleFunc("POST /api/v1/register", deps.Register)
	mux.HandleFunc("POST /api/v1/sessions", deps.CreateSession)
	mux.HandleFunc("POST /api/v1/scores", deps.SubmitScore)
	mux.HandleFunc("GET /api/v1/games/{slug}/leaderboard", deps.GetLeaderboard)
	mux.HandleFunc("GET /api/v1/users/{uuid}", deps.GetUser)
	mux.HandleFunc("PATCH /api/v1/users/{uuid}", deps.UpdateUser)
	mux.HandleFunc("GET /api/v1/avatars/{id}", deps.AvatarHandler())

	// Middleware chain: recovery → logging → per-IP rate limit → router.
	var h http.Handler = mux
	h = middleware.PerIPRateLimit(generalRL, cfg.TrustedProxies)(h)
	h = middleware.Logging(cfg.TrustedProxies)(h)
	h = middleware.Recovery(h)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      h,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		slog.Info("server starting", "port", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}
}
