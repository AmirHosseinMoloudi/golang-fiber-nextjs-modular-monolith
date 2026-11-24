package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-todo/modular-monolith/config"
	"github.com/golang-todo/modular-monolith/internal/core/observability"
	"github.com/golang-todo/modular-monolith/internal/server"
)

func main() {
	cfg := config.LoadConfig()
	logger := observability.InitLogger(cfg.Environment)
	logger.Info().Str("env", cfg.Environment).Str("port", cfg.Port).Msg("Starting Todo Modular Monolith Service")

	// Initialize OpenTelemetry
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp, err := observability.InitTracer(ctx, "todo-backend", cfg.OTelEndpoint)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to initialize tracer")
	}
	defer func() {
		if tp != nil {
			_ = tp.Shutdown(context.Background())
		}
	}()

	srv, err := server.NewServer(cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize server")
	}

	// Start Outbox background worker
	srv.OutboxDispatcher.StartWorker(ctx, 3*time.Second)

	// Graceful shutdown channel
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	go func() {
		addr := fmt.Sprintf(":%s", cfg.Port)
		logger.Info().Str("address", addr).Msg("HTTP server listening")
		if err := srv.App.Listen(addr); err != nil {
			logger.Info().Err(err).Msg("Server stopped listening")
		}
	}()

	<-shutdown
	logger.Info().Msg("Gracefully shutting down...")
	_ = srv.App.Shutdown()
	logger.Info().Msg("Server stopped")
}
