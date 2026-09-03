package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gopulse/internal/models"
)

type MetricsRepository interface {
	Insert(ctx context.Context, metric *models.RequestMetric) error
	InsertBatch(ctx context.Context, metrics []*models.RequestMetric) error
}

type metricsRepo struct {
	db *sqlx.DB
}

func NewMetricsRepository(db *sqlx.DB) MetricsRepository {
	return &metricsRepo{db: db}
}

func (r *metricsRepo) Insert(ctx context.Context, metric *models.RequestMetric) error {
	query := `
		INSERT INTO request_metrics (service_id, endpoint_id, request_id, method, path, status_code, latency_ms, client_ip, timestamp)
		VALUES (:service_id, :endpoint_id, :request_id, :method, :path, :status_code, :latency_ms, :client_ip, :timestamp)
		RETURNING id
	`
	rows, err := r.db.NamedQueryContext(ctx, query, metric)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.StructScan(metric)
	}
	return nil
}

func (r *metricsRepo) InsertBatch(ctx context.Context, metrics []*models.RequestMetric) error {
	if len(metrics) == 0 {
		return nil
	}

	query := `
		INSERT INTO request_metrics (service_id, endpoint_id, request_id, method, path, status_code, latency_ms, client_ip, timestamp)
		VALUES (:service_id, :endpoint_id, :request_id, :method, :path, :status_code, :latency_ms, :client_ip, :timestamp)
	`
	// NamedExec is efficient for batch inserts
	_, err := r.db.NamedExecContext(ctx, query, metrics)
	return err
}
