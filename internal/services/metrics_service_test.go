package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"gopulse/internal/errors"
	"gopulse/internal/models"
	"gopulse/internal/services"
)

// --- Mocks ---

type mockServiceRepo struct {
	GetServiceFunc func(apiKey string) (*models.Service, error)
}
func (m *mockServiceRepo) Create(ctx context.Context, service *models.Service) error { return nil }
func (m *mockServiceRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.Service, error) { return nil, nil }
func (m *mockServiceRepo) GetByAPIKey(ctx context.Context, apiKey string) (*models.Service, error) {
	return m.GetServiceFunc(apiKey)
}
func (m *mockServiceRepo) List(ctx context.Context, limit, offset int) ([]*models.Service, error) { return nil, nil }
func (m *mockServiceRepo) Delete(ctx context.Context, id uuid.UUID) error { return nil }

type mockEndpointRepo struct {
	GetOrCreateFunc func(serviceID uuid.UUID, path, method string) (*models.Endpoint, error)
}
func (m *mockEndpointRepo) Create(ctx context.Context, endpoint *models.Endpoint) error { return nil }
func (m *mockEndpointRepo) GetOrCreate(ctx context.Context, serviceID uuid.UUID, path, method string) (*models.Endpoint, error) {
	return m.GetOrCreateFunc(serviceID, path, method)
}
func (m *mockEndpointRepo) ListByService(ctx context.Context, serviceID uuid.UUID) ([]*models.Endpoint, error) { return nil, nil }

type mockMetricsRepo struct {
	InsertFunc func(metric *models.RequestMetric) error
}
func (m *mockMetricsRepo) Insert(ctx context.Context, metric *models.RequestMetric) error {
	return m.InsertFunc(metric)
}
func (m *mockMetricsRepo) InsertBatch(ctx context.Context, metrics []*models.RequestMetric) error { return nil }

type mockErrorRepo struct {
	InsertFunc func(errModel *models.Error) error
}
func (m *mockErrorRepo) Insert(ctx context.Context, errModel *models.Error) error {
	return m.InsertFunc(errModel)
}

type mockCache struct {}
func (m *mockCache) Get(ctx context.Context, key string) (string, error) { return "", errors.NewNotFound("not found") }
func (m *mockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error { return nil }
func (m *mockCache) Delete(ctx context.Context, key string) error { return nil }

// --- Tests ---

func TestRecordMetric_Success(t *testing.T) {
	serviceID := uuid.New()
	endpointID := uuid.New()

	serviceRepo := &mockServiceRepo{
		GetServiceFunc: func(apiKey string) (*models.Service, error) {
			if apiKey == "valid-key" {
				return &models.Service{ID: serviceID, Name: "TestService"}, nil
			}
			return nil, errors.NewUnauthorized("invalid key")
		},
	}

	endpointRepo := &mockEndpointRepo{
		GetOrCreateFunc: func(sid uuid.UUID, path, method string) (*models.Endpoint, error) {
			return &models.Endpoint{ID: endpointID, ServiceID: sid, Path: path, Method: method}, nil
		},
	}

	metricInserted := false
	metricsRepo := &mockMetricsRepo{
		InsertFunc: func(metric *models.RequestMetric) error {
			if metric.ServiceID == serviceID && metric.Path == "/test" {
				metricInserted = true
			}
			return nil
		},
	}

	svc := services.NewMetricsService(serviceRepo, endpointRepo, metricsRepo, &mockErrorRepo{}, &mockCache{})

	metric := &models.RequestMetric{
		RequestID:  "req-1",
		Method:     "GET",
		Path:       "/test",
		StatusCode: 200,
		LatencyMs:  45,
	}

	err := svc.RecordMetric(context.Background(), "valid-key", metric)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !metricInserted {
		t.Errorf("expected metric to be inserted")
	}
}

func TestRecordMetric_InvalidKey(t *testing.T) {
	serviceRepo := &mockServiceRepo{
		GetServiceFunc: func(apiKey string) (*models.Service, error) {
			return nil, errors.NewUnauthorized("invalid key")
		},
	}

	svc := services.NewMetricsService(serviceRepo, &mockEndpointRepo{}, &mockMetricsRepo{}, &mockErrorRepo{}, &mockCache{})

	err := svc.RecordMetric(context.Background(), "invalid-key", &models.RequestMetric{})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	
	appErr, ok := err.(*errors.AppError)
	if !ok || appErr.Code != errors.Unauthorized {
		t.Errorf("expected Unauthorized error, got %v", err)
	}
}
