package grpc

import (
	"context"
	"errors"
	"log/slog"

	pb "github.com/Pavel26ru/aggregator-service/gen"
	"github.com/Pavel26ru/aggregator-service/internal/repository"
	"github.com/Pavel26ru/aggregator-service/internal/service"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	pb.UnimplementedAggregatorServiceServer
	service *service.Service
	log     *slog.Logger
}

func NewHandler(s *service.Service, log *slog.Logger) *Handler {
	return &Handler{service: s, log: log}
}

func (h *Handler) GetMax(ctx context.Context, req *pb.GetMaxRequest) (*pb.GetMaxResponse, error) {
	const op = "grpc.GetMax"
	log := h.log.With(slog.String("op", op), slog.Any("request", req))

	// по UUID
	if req.Uuid != "" {
		rec, err := h.service.GetMaxByID(ctx, req.Uuid)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				log.Info("record not found by uuid")
				return nil, status.Error(codes.NotFound, "record not found")
			}
			log.Error("failed to get record by id", slog.Any("error", err))
			return nil, status.Error(codes.Internal, "internal error")
		}

		return &pb.GetMaxResponse{
			Records: []*pb.MaxValue{
				{
					MaxValue: rec.Value,
				},
			},
		}, nil
	}

	// по периоду
	if req.From != nil && req.To != nil {
		from := req.From.AsTime()
		to := req.To.AsTime()

		list, err := h.service.GetMaxByPeriod(ctx, from, to)
		if err != nil {
			log.Error("failed to get records by period", slog.Any("error", err))
			return nil, status.Error(codes.Internal, "internal error")
		}
		if len(list) == 0 {
			log.Info("no records found for the period")
			return &pb.GetMaxResponse{}, nil
		}

		resp := &pb.GetMaxResponse{}
		for _, rec := range list {
			resp.Records = append(resp.Records, &pb.MaxValue{
				MaxValue: rec.Value,
			})
		}
		return resp, nil
	}

	log.Warn("bad request: neither uuid nor period provided")
	return nil, status.Error(codes.InvalidArgument, "either uuid or a time period must be provided")
}
