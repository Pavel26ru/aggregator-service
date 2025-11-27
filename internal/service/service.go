package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/Pavel26ru/aggregator-service/internal/model"
	"github.com/Pavel26ru/aggregator-service/internal/repository"
)

type Service struct {
	logger  *slog.Logger
	pgxrepo repository.MaxValueRepository
}

func New(logger *slog.Logger, pgxrepo repository.MaxValueRepository) *Service {
	return &Service{logger: logger, pgxrepo: pgxrepo}
}

func (s *Service) SaveMaxValue(ctx context.Context, rec model.MaxValueRecord) error {
	return s.pgxrepo.SaveMax(ctx, &rec)
}

func (s *Service) GetMaxByID(ctx context.Context, uuid string) (*model.MaxValue, error) {
	return s.pgxrepo.GetMaxByID(ctx, uuid)
}

func (s *Service) GetMaxByPeriod(ctx context.Context, from, to time.Time) ([]model.MaxValue, error) {
	return s.pgxrepo.GetMaxByPeriod(ctx, from, to)
}

func (s *Service) ComputeMax(values []int64) int64 {
	if len(values) == 0 {
		return 0
	}
	_max_ := values[0]
	for _, v := range values[1:] {
		if v > _max_ {
			_max_ = v
		}
	}
	return _max_
}
