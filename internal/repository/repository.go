package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Pavel26ru/aggregator-service/internal/model"
)

var ErrNotFound = errors.New("not found")

type MaxValueRepository interface {
	SaveMax(ctx context.Context, rec *model.MaxValueRecord) error
	GetMaxByID(ctx context.Context, uuid string) (*model.MaxValue, error)
	GetMaxByPeriod(ctx context.Context, from, to time.Time) ([]model.MaxValue, error)
}
