package kafka

import (
	"context"
	"errors"
	"fmt"

	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"
)

// EnsureTopic создаёт Kafka topic, если он отсутствует, используя franz-go.
func EnsureTopic(ctx context.Context, brokerAddr, topic string, partitions int) error {
	client, err := kgo.NewClient(kgo.SeedBrokers(brokerAddr))
	if err != nil {
		return fmt.Errorf("failed to create kafka client: %w", err)
	}
	defer client.Close()

	req := &kmsg.CreateTopicsRequest{
		Topics: []kmsg.CreateTopicsRequestTopic{
			{
				Topic:             topic,
				NumPartitions:     int32(partitions),
				ReplicationFactor: 1,
			},
		},
		TimeoutMillis: 5000,
	}

	rawResp, err := client.Request(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to request topic creation: %w", err)
	}

	resp, ok := rawResp.(*kmsg.CreateTopicsResponse)
	if !ok {
		return fmt.Errorf("unexpected response type: %T", rawResp)
	}

	if len(resp.Topics) != 1 {
		return fmt.Errorf("unexpected response length: %d", len(resp.Topics))
	}
	topicResp := resp.Topics[0]

	if err := kerr.ErrorForCode(topicResp.ErrorCode); err != nil && !errors.Is(err, kerr.TopicAlreadyExists) {
		return fmt.Errorf("failed to create topic '%s': %w", topic, err)
	}

	return nil
}
