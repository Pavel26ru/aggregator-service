package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/Pavel26ru/aggregator-service/internal/model"
	"github.com/Pavel26ru/aggregator-service/internal/repository"
	"github.com/Pavel26ru/aggregator-service/internal/service/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_GetMaxByID(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	testUUID := "test-uuid-123"
	expectedRecord := &model.MaxValue{
		Value: 100,
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo := &mocks.MockMaxValueRepository{
			GetMaxByIDFunc: func(ctx context.Context, uuid string) (*model.MaxValue, error) {
				assert.Equal(t, testUUID, uuid)
				return expectedRecord, nil
			},
		}
		service := New(logger, mockRepo)

		record, err := service.GetMaxByID(ctx, testUUID)

		require.NoError(t, err)
		assert.Equal(t, expectedRecord, record)
	})

	t.Run("Not Found", func(t *testing.T) {
		mockRepo := &mocks.MockMaxValueRepository{
			GetMaxByIDFunc: func(ctx context.Context, uuid string) (*model.MaxValue, error) {
				assert.Equal(t, testUUID, uuid)
				return nil, repository.ErrNotFound
			},
		}
		service := New(logger, mockRepo)

		record, err := service.GetMaxByID(ctx, testUUID)

		require.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrNotFound)
		assert.Nil(t, record)
	})
}

func TestService_GetMaxByPeriod(t *testing.T) {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	from := time.Now().Add(-1 * time.Hour)
	to := time.Now()
	expectedRecords := []model.MaxValue{
		{Value: 100},
		{Value: 200},
	}

	t.Run("Success", func(t *testing.T) {
		mockRepo := &mocks.MockMaxValueRepository{
			GetMaxByPeriodFunc: func(ctx context.Context, fromTime, toTime time.Time) ([]model.MaxValue, error) {
				assert.WithinDuration(t, from, fromTime, time.Second)
				assert.WithinDuration(t, to, toTime, time.Second)
				return expectedRecords, nil
			},
		}
		service := New(logger, mockRepo)

		records, err := service.GetMaxByPeriod(ctx, from, to)

		require.NoError(t, err)
		assert.Equal(t, expectedRecords, records)
	})

	t.Run("Empty result", func(t *testing.T) {
		mockRepo := &mocks.MockMaxValueRepository{
			GetMaxByPeriodFunc: func(ctx context.Context, f, t time.Time) ([]model.MaxValue, error) {
				return []model.MaxValue{}, nil
			},
		}
		service := New(logger, mockRepo)

		records, err := service.GetMaxByPeriod(ctx, from, to)

		require.NoError(t, err)
		assert.Empty(t, records)
	})
}
