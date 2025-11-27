package ingestion

import (
	"context"
	"log/slog"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"github.com/Pavel26ru/aggregator-service/internal/kafka"
	"github.com/Pavel26ru/aggregator-service/internal/model"
)

type Generator struct {
	interval time.Duration
	producer kafka.Producer
	log      *slog.Logger
	r        *rand.Rand
}

func NewGenerator(interval time.Duration, producer kafka.Producer, log *slog.Logger) *Generator {
	return &Generator{
		interval: interval,
		producer: producer,
		log:      log.With("component", "generator"),
		r:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *Generator) Start(ctx context.Context) {
	ticker := time.NewTicker(g.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			g.log.Info("generator stopped")
			return

		case <-ticker.C:
			rec := model.ValueRecord{
				UUID:      uuid.New().String(),
				Value:     g.generateRandomSlice(),
				Timestamp: time.Now().UTC(),
			}

			if err := g.producer.Produce(ctx, rec); err != nil {
				g.log.Error("failed to produce record", slog.Any("error", err))
			}
		}
	}
}

func (g *Generator) generateRandomSlice() []int64 {
	// Генерируем срез случайной длины от 5 до 15
	sliceLen := g.r.Intn(11) + 5
	slice := make([]int64, sliceLen)
	for i := 0; i < sliceLen; i++ {
		slice[i] = g.r.Int63()
	}
	return slice
}
