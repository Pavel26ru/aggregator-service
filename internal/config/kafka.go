package config

type KafkaConfig struct {
	Brokers []string
	Topic   string
	Group   string
}
