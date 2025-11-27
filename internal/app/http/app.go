package httpapp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/Pavel26ru/aggregator-service/internal/service"
	"github.com/Pavel26ru/aggregator-service/internal/transport/rest"
)

type App struct {
	httpServer *http.Server
	logger     *slog.Logger
	address    string
}

func New(ctx context.Context, logger *slog.Logger, s *service.Service, address string) *App {
	log := logger.With(slog.String("component", "httpapp"))

	router := rest.New(s, log)

	httpServer := &http.Server{
		Addr:    address,
		Handler: router,
	}

	return &App{
		httpServer: httpServer,
		logger:     log,
		address:    address,
	}
}

func (a *App) Run() error {
	const op = "httpapp.Run"
	log := a.logger.With(slog.String("op", op))

	log.Info("http server is running", slog.String("address", a.address))

	if err := a.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (a *App) Stop(ctx context.Context) error {
	const op = "httpapp.Stop"
	log := a.logger.With(slog.String("op", op))

	log.Info("stopping http server")

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}