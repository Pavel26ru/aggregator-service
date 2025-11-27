package mocks

import (
	"context"
	"time"

	"github.com/Pavel26ru/aggregator-service/internal/model"
)

// MockMaxValueRepository is a mock implementation of the MaxValueRepository interface.
// It allows for setting expected return values for testing purposes.
type MockMaxValueRepository struct {
	SaveMaxFunc        func(ctx context.Context, rec *model.MaxValueRecord) error
	GetMaxByIDFunc     func(ctx context.Context, uuid string) (*model.MaxValue, error)
	GetMaxByPeriodFunc func(ctx context.Context, from, to time.Time) ([]model.MaxValue, error)
}

func (m *MockMaxValueRepository) SaveMax(ctx context.Context, rec *model.MaxValueRecord) error {
	if m.SaveMaxFunc != nil {
		return m.SaveMaxFunc(ctx, rec)
	}
	return nil
}

func (m *MockMaxValueRepository) GetMaxByID(ctx context.Context, uuid string) (*model.MaxValue, error) {
	if m.GetMaxByIDFunc != nil {
		return m.GetMaxByIDFunc(ctx, uuid)
	}
	return nil, nil
}

func (m *MockMaxValueRepository) GetMaxByPeriod(ctx context.Context, from, to time.Time) ([]model.MaxValue, error) {
	if m.GetMaxByPeriodFunc != nil {
		return m.GetMaxByPeriodFunc(ctx, from, to)
	}
	return nil, nil
}
