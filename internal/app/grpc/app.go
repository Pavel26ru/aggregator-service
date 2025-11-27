package grpcapp

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	grpcMaxValue "github.com/Pavel26ru/aggregator-service/gen"
	"github.com/Pavel26ru/aggregator-service/internal/service"
	grpchandler "github.com/Pavel26ru/aggregator-service/internal/transport/grpc"
	"google.golang.org/grpc"
)

type App struct {
	gRPCServer *grpc.Server
	logger     *slog.Logger
	address    string
}

func New(ctx context.Context, logger *slog.Logger, s *service.Service, address string) *App {
	gRPCServer := grpc.NewServer()
	handler := grpchandler.NewHandler(s, logger)
	grpcMaxValue.RegisterAggregatorServiceServer(gRPCServer, handler)
	return &App{gRPCServer: gRPCServer, logger: logger, address: address}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	log := a.logger.With(slog.String("op", op))

	l, err := net.Listen("tcp", a.address)
	if err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	log.Info("grpc server auth is running", slog.String("address", a.address))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s:%w", op, err)
	}

	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	log := a.logger.With(slog.String("op", op))
	log.Info("stopping gRPC server")

	a.gRPCServer.GracefulStop()
}
