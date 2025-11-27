package postgres

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Pavel26ru/aggregator-service/internal/config"
	"github.com/Pavel26ru/aggregator-service/internal/model"
	"github.com/Pavel26ru/aggregator-service/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Database struct {
	db  *pgxpool.Pool
	log *slog.Logger
}

func New(ctx context.Context, cfg config.PostgresConfig, log *slog.Logger) (*Database, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DBName,
		cfg.SSLMode,
	)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Database{db: pool, log: log}, nil
}

func (d *Database) Close() {
	d.log.Info("closing postgres connection pool")
	d.db.Close()
}

func (d *Database) SaveMax(ctx context.Context, rec *model.MaxValueRecord) error {
	const q = `
		INSERT INTO max_values (uuid, ts, max_value)
		VALUES ($1, $2, $3)
		ON CONFLICT (uuid) DO UPDATE SET
			ts = EXCLUDED.ts,
			max_value = EXCLUDED.max_value;
	`

	if _, err := d.db.Exec(ctx, q, rec.UUID, rec.Timestamp, rec.MaxValue); err != nil {
		d.log.Error("SaveMax failed", slog.Any("error", err))
		return err
	}
	return nil
}

func (d *Database) GetMaxByID(ctx context.Context, uuid string) (*model.MaxValue, error) {
	const q = `
		SELECT max_value
		FROM max_values
		WHERE uuid = $1
	`

	row := d.db.QueryRow(ctx, q, uuid)

	var rec model.MaxValue
	err := row.Scan(&rec.Value)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		d.log.Error("GetMaxByID failed", slog.Any("error", err))
		return nil, err
	}

	return &rec, nil
}

func (r *Database) GetMaxByPeriod(ctx context.Context, from, to time.Time) ([]model.MaxValue, error) {
	const q = `
		SELECT max_value
		FROM max_values
		WHERE ts >= $1 AND ts <= $2
		ORDER BY ts ASC
	`

	rows, err := r.db.Query(ctx, q, from, to)
	if err != nil {
		r.log.Error("GetMaxByPeriod failed", slog.Any("error", err))
		return nil, err
	}
	defer rows.Close()

	var records []model.MaxValue
	for rows.Next() {
		var rec model.MaxValue
		if err := rows.Scan(&rec.Value); err != nil {
			return nil, err
		}
		records = append(records, rec)
	}
	if err := rows.Err(); err != nil {
		r.log.Error("GetMaxByPeriod row iteration failed", slog.Any("error", err))
		return nil, err
	}

	return records, nil
}
