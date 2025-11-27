package kafka

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/Pavel26ru/aggregator-service/internal/model"
	"github.com/Pavel26ru/aggregator-service/internal/service"
	"github.com/twmb/franz-go/pkg/kgo"
)

type FranzConsumer struct {
	client  *kgo.Client
	log     *slog.Logger
	service *service.Service
}

func NewConsumer(brokers []string, group, topic string, svc *service.Service, log *slog.Logger) (Consumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topic),
	)
	if err != nil {
		return nil, err
	}

	return &FranzConsumer{
		client:  client,
		log:     log.With("component", "kafka_consumer"),
		service: svc,
	}, nil
}

func (c *FranzConsumer) Run(ctx context.Context) error {
	for {
		fetches := c.client.PollFetches(ctx)

		if errs := fetches.Errors(); len(errs) > 0 {
			for _, e := range errs {
				c.log.Error("poll error", slog.Any("error", e))
			}
			continue
		}

		fetches.EachRecord(func(r *kgo.Record) {
			var msg model.ValueRecord

			if err := json.Unmarshal(r.Value, &msg); err != nil {
				c.log.Error("decode error", slog.Any("error", err))
				return
			}

			maxVal := c.service.ComputeMax(msg.Value)

			err := c.service.SaveMaxValue(ctx, model.MaxValueRecord{
				UUID:      msg.UUID,
				Timestamp: msg.Timestamp,
				MaxValue:  maxVal,
			})

			if err != nil {
				c.log.Error("failed to save max record", slog.Any("error", err))
			}
		})
	}
}

func (c *FranzConsumer) Close() {
	c.client.Close()
}
