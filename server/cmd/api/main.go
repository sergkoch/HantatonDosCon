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

	"accessible-content/server/internal/api"
	"accessible-content/server/internal/config"
	"accessible-content/server/internal/crawler"
	"accessible-content/server/internal/store"
)

func main() {
	cfg := config.Load()
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	memory := store.NewMemory()
	handler := api.NewRouter(api.Options{
		Store:        memory,
		Log:          log,
		AIConfigured: cfg.AIKey != "",
		Crawler:      crawler.New(crawler.Options{}),
	})

	server := &http.Server{
		Addr:              cfg.Listen,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Info("api started", "listen", cfg.Listen)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("api stopped", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("shutdown failed", "error", err)
	}
}
