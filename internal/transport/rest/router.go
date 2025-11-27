package rest

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Pavel26ru/aggregator-service/internal/metrics"
	"github.com/Pavel26ru/aggregator-service/internal/repository"
	"github.com/Pavel26ru/aggregator-service/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handler struct {
	service *service.Service
	log     *slog.Logger
}

func New(s *service.Service, log *slog.Logger) *chi.Mux {
	r := chi.NewRouter()
	h := &Handler{service: s, log: log}

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(AccessLogMiddleware(log))
	r.Use(metrics.Middleware)

	r.Handle("/metrics", promhttp.Handler())
	r.Get("/max", h.GetMax)

	return r
}

func (h *Handler) GetMax(w http.ResponseWriter, r *http.Request) {
	const op = "rest.GetMax"
	log := h.log.With(slog.String("op", op))

	uuid := r.URL.Query().Get("uuid")
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")

	// По ID
	if uuid != "" {
		rec, err := h.service.GetMaxByID(r.Context(), uuid)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				log.Info("record not found", slog.String("uuid", uuid))
				http.Error(w, "record not found", http.StatusNotFound)
				return
			}
			log.Error("failed to get record by id", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
		respondJSON(w, http.StatusOK, rec)
		return
	}

	// По периоду
	if fromStr != "" && toStr != "" {
		from, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			log.Error("invalid 'from' timestamp format", slog.Any("error", err))
			http.Error(w, "invalid 'from' timestamp format", http.StatusBadRequest)
			return
		}

		to, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			log.Error("invalid 'to' timestamp format", slog.Any("error", err))
			http.Error(w, "invalid 'to' timestamp format", http.StatusBadRequest)
			return
		}

		list, err := h.service.GetMaxByPeriod(r.Context(), from, to)
		if err != nil {
			log.Error("failed to get records by period", slog.Any("error", err))
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}

		respondJSON(w, http.StatusOK, list)
		return
	}

	http.Error(w, "bad request: either uuid or a time period must be provided", http.StatusBadRequest)
}

func respondJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		// Логируем ошибку, но уже не можем отправить другой статус, т.к. заголовок записан
		slog.Default().Error("failed to encode json response", slog.Any("error", err))
	}
}
