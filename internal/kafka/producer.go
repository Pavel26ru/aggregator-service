package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Pavel26ru/aggregator-service/internal/model"
	"github.com/twmb/franz-go/pkg/kgo"
)

type producer struct {
	client *kgo.Client
	topic  string
	log    *slog.Logger
}

func NewProducer(brokers []string, topic string, log *slog.Logger) (Producer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
	)
	if err != nil {
		return nil, err
	}

	return &producer{
		client: client,
		topic:  topic,
		log:    log.With("component", "kafka_producer"),
	}, nil
}

func (p *producer) Produce(ctx context.Context, rec model.ValueRecord) error {
	value, err := json.Marshal(rec)
	if err != nil {
		return err
	}

	record := &kgo.Record{
		Topic: p.topic,
		Value: value,
		Key:   []byte(rec.UUID),
	}

	p.client.Produce(ctx, record, func(_ *kgo.Record, err error) {
		if err != nil {
			p.log.Error("failed to deliver record", slog.Any("error", err))
		}
	})

	return nil
}

func (p *producer) Close() {
	p.client.Close()
}
