package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Pavel26ru/aggregator-service/internal/app"
	"github.com/Pavel26ru/aggregator-service/internal/config"
	"github.com/Pavel26ru/aggregator-service/internal/logging"
	"github.com/Pavel26ru/aggregator-service/internal/shutdown"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM,
	)
	defer cancel()

	if err := Run(ctx); err != nil {
		log.Printf("service stopped with error: %v", err)
		os.Exit(1)
	}

	log.Println("service stopped gracefully")
}

func Run(ctx context.Context) error {
	cfg := config.Load()

	logger := logging.SetupLogger(cfg.Env).With(
		slog.String("component", "app"),
	)

	sh := shutdown.New(logger)

	application := app.New(ctx, logger, cfg)

	sh.Register(application.Stop)

	logger.Info("starting aggregator service",
		slog.String("env", cfg.Env),
		slog.String("rest", cfg.HTTP.Port),
		slog.String("gRPC", cfg.GRPC.Port),
	)

	<-ctx.Done()
	logger.Info("received shutdown signal")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	sh.Shutdown(shutdownCtx)

	return nil
}
