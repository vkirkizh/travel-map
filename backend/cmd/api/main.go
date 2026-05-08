package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/vkirkizh/travel-map/backend/internal/config"
	"github.com/vkirkizh/travel-map/backend/internal/server"
	"github.com/vkirkizh/travel-map/backend/internal/storage"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := storage.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           server.New(db, cfg),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("starting http server", "addr", cfg.HTTPAddr, "env", cfg.AppEnv)

		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "error", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("shutting down http server")

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown http server gracefully", "error", err)
		os.Exit(1)
	}

	slog.Info("server stopped")
}
