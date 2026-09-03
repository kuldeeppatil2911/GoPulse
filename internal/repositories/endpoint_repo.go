package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"gopulse/internal/models"
)

type EndpointRepository interface {
	Create(ctx context.Context, endpoint *models.Endpoint) error
	GetOrCreate(ctx context.Context, serviceID uuid.UUID, path, method string) (*models.Endpoint, error)
	ListByService(ctx context.Context, serviceID uuid.UUID) ([]*models.Endpoint, error)
}

type endpointRepo struct {
	db *sqlx.DB
}

func NewEndpointRepository(db *sqlx.DB) EndpointRepository {
	return &endpointRepo{db: db}
}

func (r *endpointRepo) Create(ctx context.Context, endpoint *models.Endpoint) error {
	query := `
		INSERT INTO endpoints (service_id, path, method) 
		VALUES (:service_id, :path, :method) 
		RETURNING id, created_at
	`
	rows, err := r.db.NamedQueryContext(ctx, query, endpoint)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		return rows.StructScan(endpoint)
	}
	return nil
}

func (r *endpointRepo) GetOrCreate(ctx context.Context, serviceID uuid.UUID, path, method string) (*models.Endpoint, error) {
	var endpoint models.Endpoint
	// Try to get first
	err := r.db.GetContext(ctx, &endpoint, "SELECT * FROM endpoints WHERE service_id = $1 AND path = $2 AND method = $3", serviceID, path, method)

	if err == nil {
		return &endpoint, nil
	}

	// If not found, create
	newEndpoint := &models.Endpoint{
		ServiceID: serviceID,
		Path:      path,
		Method:    method,
	}

	// Use ON CONFLICT to handle race conditions gracefully
	query := `
		INSERT INTO endpoints (service_id, path, method) 
		VALUES ($1, $2, $3) 
		ON CONFLICT (service_id, path, method) DO UPDATE SET path=EXCLUDED.path
		RETURNING id, created_at
	`
	err = r.db.QueryRowContext(ctx, query, serviceID, path, method).Scan(&newEndpoint.ID, &newEndpoint.CreatedAt)
	if err != nil {
		return nil, err
	}

	return newEndpoint, nil
}

func (r *endpointRepo) ListByService(ctx context.Context, serviceID uuid.UUID) ([]*models.Endpoint, error) {
	var endpoints []*models.Endpoint
	err := r.db.SelectContext(ctx, &endpoints, "SELECT * FROM endpoints WHERE service_id = $1 ORDER BY path ASC", serviceID)
	return endpoints, err
}
