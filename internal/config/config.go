package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env      string
	HTTP     HTTPConfig
	GRPC     GRPCConfig
	Postgres PostgresConfig
	Kafka    KafkaConfig

	Workers  int
	Interval time.Duration
}

func Load() *Config {
	return &Config{
		Env: getEnv("ENV", "local"),

		HTTP: HTTPConfig{
			Port: getEnv("HTTP_PORT", "8080"),
		},

		GRPC: GRPCConfig{
			Port: getEnv("GRPC_PORT", "9090"),
		},

		Postgres: PostgresConfig{
			Host:     getEnv("POSTGRES_HOST", "postgres"),
			Port:     getEnvInt("POSTGRES_PORT", 5432),
			User:     getEnv("POSTGRES_USER", "postgres"),
			Password: getEnv("POSTGRES_PASSWORD", "postgres"),
			DBName:   getEnv("POSTGRES_DB", "agg"),
			SSLMode:  getEnv("POSTGRES_SSL", "disable"),
		},

		Kafka: KafkaConfig{
			Brokers: parseList(getEnv("KAFKA_BROKERS", "kafka:9092")),
			Topic:   getEnv("KAFKA_TOPIC", "records"),
			Group:   getEnv("KAFKA_GROUP", "agg-workers"),
		},

		Workers:  getEnvInt("WORKERS", 5),
		Interval: getEnvDuration("INTERVAL", "100ms"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		parsed, err := strconv.Atoi(val)
		if err != nil {
			log.Printf("invalid int for %s: %s, using default %d", key, val, defaultVal)
			return defaultVal
		}
		return parsed
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal string) time.Duration {
	if val, ok := os.LookupEnv(key); ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			log.Printf("invalid duration for %s: %s, using default %s", key, val, defaultVal)
			parsed, _ := time.ParseDuration(defaultVal)
			return parsed
		}
		return d
	}
	d, _ := time.ParseDuration(defaultVal)
	return d
}

func parseList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
