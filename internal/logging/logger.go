package logging

import (
	"log/slog"
	"os"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func SetupLogger(env string) *slog.Logger {
	var handler slog.Handler

	switch env {
	case envProd:
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	default:
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	}

	return slog.New(handler)
}
