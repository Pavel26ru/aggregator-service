package kafka

import (
	"context"
	"github.com/Pavel26ru/aggregator-service/internal/model"
)

type Producer interface {
	Produce(ctx context.Context, rec model.ValueRecord) error
	Close()
}

type Consumer interface {
	Run(ctx context.Context) error
	Close()
}
