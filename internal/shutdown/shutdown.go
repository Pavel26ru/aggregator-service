package shutdown

import (
	"context"
	"log/slog"
	"sync"
)

type Hook func(ctx context.Context) error

type Manager struct {
	log   *slog.Logger
	hooks []Hook
	mu    sync.Mutex
}

func New(log *slog.Logger) *Manager {
	return &Manager{
		log: log.With(slog.String("component", "shutdown")),
	}
}

func (m *Manager) Register(h Hook) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks = append(m.hooks, h)
}

func (m *Manager) Shutdown(ctx context.Context) {
	m.mu.Lock()
	hooks := make([]Hook, len(m.hooks))
	copy(hooks, m.hooks)
	m.mu.Unlock()

	m.log.Info("starting graceful shutdown")

	for i := len(hooks) - 1; i >= 0; i-- {
		if err := hooks[i](ctx); err != nil {
			m.log.Error("shutdown hook failed", slog.Any("error", err))
		}
	}

	m.log.Info("graceful shutdown complete")
}
