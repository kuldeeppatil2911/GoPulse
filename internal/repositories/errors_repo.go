package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
	"gopulse/internal/models"
)

type ErrorRepository interface {
	Insert(ctx context.Context, errModel *models.Error) error
}

type errorRepo struct {
	db *sqlx.DB
}

func NewErrorRepository(db *sqlx.DB) ErrorRepository {
	return &errorRepo{db: db}
}

func (r *errorRepo) Insert(ctx context.Context, errModel *models.Error) error {
	query := `
		INSERT INTO errors (service_id, request_id, error_code, error_message, stack_trace, timestamp)
		VALUES (:service_id, :request_id, :error_code, :error_message, :stack_trace, :timestamp)
		RETURNING id
	`
	rows, err := r.db.NamedQueryContext(ctx, query, errModel)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.StructScan(errModel)
	}
	return nil
}
