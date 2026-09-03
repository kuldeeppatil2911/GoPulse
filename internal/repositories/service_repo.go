package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gopulse/internal/models"
)

type ServiceRepository interface {
	Create(ctx context.Context, service *models.Service) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.Service, error)
	GetByAPIKey(ctx context.Context, apiKey string) (*models.Service, error)
	List(ctx context.Context, limit, offset int) ([]*models.Service, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type serviceRepo struct {
	db *sqlx.DB
}

func NewServiceRepository(db *sqlx.DB) ServiceRepository {
	return &serviceRepo{db: db}
}

func (r *serviceRepo) Create(ctx context.Context, service *models.Service) error {
	query := `
		INSERT INTO services (name, description, api_key) 
		VALUES (:name, :description, :api_key) 
		RETURNING id, created_at, updated_at
	`
	rows, err := r.db.NamedQueryContext(ctx, query, service)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.StructScan(service)
	}
	return nil
}

func (r *serviceRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Service, error) {
	var service models.Service
	err := r.db.GetContext(ctx, &service, "SELECT * FROM services WHERE id = $1", id)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepo) GetByAPIKey(ctx context.Context, apiKey string) (*models.Service, error) {
	var service models.Service
	err := r.db.GetContext(ctx, &service, "SELECT * FROM services WHERE api_key = $1", apiKey)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (r *serviceRepo) List(ctx context.Context, limit, offset int) ([]*models.Service, error) {
	var services []*models.Service
	err := r.db.SelectContext(ctx, &services, "SELECT * FROM services ORDER BY created_at DESC LIMIT $1 OFFSET $2", limit, offset)
	return services, err
}

func (r *serviceRepo) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM services WHERE id = $1", id)
	return err
}
