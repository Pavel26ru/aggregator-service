package app

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/Pavel26ru/aggregator-service/internal/app/grpc"
	"github.com/Pavel26ru/aggregator-service/internal/app/http"
	"github.com/Pavel26ru/aggregator-service/internal/config"
	"github.com/Pavel26ru/aggregator-service/internal/ingestion"
	"github.com/Pavel26ru/aggregator-service/internal/kafka"
	"github.com/Pavel26ru/aggregator-service/internal/repository/postgres"
	"github.com/Pavel26ru/aggregator-service/internal/service"
)

type App struct {
	GRPCServer    *grpcapp.App
	HTTPServer    *httpapp.App
	KafkaProducer kafka.Producer
	consumers     []kafka.Consumer
	generator     *ingestion.Generator
	db            *postgres.Database
	logger        *slog.Logger
}

func New(ctx context.Context, logger *slog.Logger, cfg *config.Config) *App {
	log := logger.With(slog.String("component", "app"))

	// === DB ===
	db, err := postgres.New(ctx, cfg.Postgres, log)
	if err != nil {
		panic(fmt.Errorf("failed to init db: %w", err))
	}

	// === Service ===
	aggregatorService := service.New(log, db)

	// === Kafka Topic ===
	if err := kafka.EnsureTopic(ctx, cfg.Kafka.Brokers[0], cfg.Kafka.Topic, 10); err != nil {
		panic(fmt.Errorf("failed to ensure kafka topic: %w", err))
	}
	log.Info("kafka topic ensured", slog.String("topic", cfg.Kafka.Topic))

	// === Kafka Producer ===
	producer, err := kafka.NewProducer(cfg.Kafka.Brokers, cfg.Kafka.Topic, log)
	if err != nil {
		panic(fmt.Errorf("failed to create kafka producer: %w", err))
	}

	// === Generator ===
	generator := ingestion.NewGenerator(cfg.Interval, producer, log)
	go func() {
		log.Info("starting generator")
		generator.Start(ctx)
	}()

	// === Kafka Consumers ===
	var consumers []kafka.Consumer
	for i := 0; i < cfg.Workers; i++ {
		consumerLog := log.With(slog.Int("worker_id", i+1))
		consumer, err := kafka.NewConsumer(
			cfg.Kafka.Brokers,
			cfg.Kafka.Group,
			cfg.Kafka.Topic,
			aggregatorService,
			consumerLog,
		)
		if err != nil {
			panic(fmt.Errorf("failed to create kafka consumer worker %d: %w", i+1, err))
		}
		consumers = append(consumers, consumer)

		go func() {
			consumerLog.Info("starting consumer worker")
			if err := consumer.Run(ctx); err != nil {
				consumerLog.Error("consumer worker failed", slog.Any("error", err))
			}
		}()
	}

	// === Servers ===
	grpcApp := grpcapp.New(ctx, log, aggregatorService, cfg.GRPC.Addr())
	go func() {
		if err := grpcApp.Run(); err != nil {
			log.Error("gRPC server failed", slog.Any("error", err))
		}
	}()

	httpApp := httpapp.New(ctx, log, aggregatorService, cfg.HTTP.Addr())
	go func() {
		if err := httpApp.Run(); err != nil {
			log.Error("http server failed", slog.Any("error", err))
		}
	}()

	return &App{
		GRPCServer:    grpcApp,
		HTTPServer:    httpApp,
		KafkaProducer: producer,
		consumers:     consumers,
		generator:     generator,
		db:            db,
		logger:        log,
	}
}

func (a *App) Stop(ctx context.Context) error {
	a.logger.Info("stopping application components")

	var wg sync.WaitGroup
	var errs []error

	// Stop servers
	wg.Add(2)
	go func() {
		defer wg.Done()
		a.GRPCServer.Stop()
	}()
	go func() {
		defer wg.Done()
		if err := a.HTTPServer.Stop(ctx); err != nil {
			a.logger.Error("failed to stop http server", slog.Any("error", err))
			errs = append(errs, err)
		}
	}()

	// Stop Kafka producer
	a.KafkaProducer.Close()

	// Stop Kafka consumers
	for _, c := range a.consumers {
		c.Close()
	}

	// Stop DB
	a.db.Close()

	wg.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("shutdown finished with errors: %v", errs)
	}

	return nil
}