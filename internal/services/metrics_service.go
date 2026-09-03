package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gopulse/internal/cache"
	"gopulse/internal/errors"
	"gopulse/internal/logger"
	"gopulse/internal/models"
	"gopulse/internal/repositories"
)

type MetricsService interface {
	RecordMetric(ctx context.Context, apiKey string, metric *models.RequestMetric) error
	RecordError(ctx context.Context, apiKey string, errModel *models.Error) error
}

type metricsService struct {
	serviceRepo  repositories.ServiceRepository
	endpointRepo repositories.EndpointRepository
	metricsRepo  repositories.MetricsRepository
	errorRepo    repositories.ErrorRepository
	cache        cache.Cache
}

func NewMetricsService(
	sr repositories.ServiceRepository,
	er repositories.EndpointRepository,
	mr repositories.MetricsRepository,
	errRepo repositories.ErrorRepository,
	c cache.Cache,
) MetricsService {
	return &metricsService{
		serviceRepo:  sr,
		endpointRepo: er,
		metricsRepo:  mr,
		errorRepo:    errRepo,
		cache:        c,
	}
}

func (s *metricsService) getServiceByAPIKey(ctx context.Context, apiKey string) (*models.Service, error) {
	cacheKey := fmt.Sprintf("service:apikey:%s", apiKey)

	// Try cache first
	if val, err := s.cache.Get(ctx, cacheKey); err == nil && val != "" {
		var service models.Service
		if json.Unmarshal([]byte(val), &service) == nil {
			return &service, nil
		}
	}

	// Fetch from DB
	service, err := s.serviceRepo.GetByAPIKey(ctx, apiKey)
	if err != nil {
		return nil, errors.NewUnauthorized("Invalid API Key")
	}

	// Set cache
	if serviceBytes, err := json.Marshal(service); err == nil {
		_ = s.cache.Set(ctx, cacheKey, string(serviceBytes), 5*time.Minute)
	}

	return service, nil
}

func (s *metricsService) RecordMetric(ctx context.Context, apiKey string, metric *models.RequestMetric) error {
	service, err := s.getServiceByAPIKey(ctx, apiKey)
	if err != nil {
		return err
	}
	metric.ServiceID = service.ID

	endpoint, err := s.endpointRepo.GetOrCreate(ctx, service.ID, metric.Path, metric.Method)
	if err != nil {
		logger.Log.Error("Failed to get or create endpoint", zap.Error(err))
		// We still record the metric, just without endpoint_id
	} else {
		metric.EndpointID = &endpoint.ID
	}

	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now()
	}

	err = s.metricsRepo.Insert(ctx, metric)
	if err != nil {
		logger.Log.Error("Failed to insert metric", zap.Error(err))
		return errors.NewInternal(err)
	}

	return nil
}

func (s *metricsService) RecordError(ctx context.Context, apiKey string, errModel *models.Error) error {
	service, err := s.getServiceByAPIKey(ctx, apiKey)
	if err != nil {
		return err
	}
	errModel.ServiceID = service.ID

	if errModel.Timestamp.IsZero() {
		errModel.Timestamp = time.Now()
	}

	err = s.errorRepo.Insert(ctx, errModel)
	if err != nil {
		logger.Log.Error("Failed to insert error", zap.Error(err))
		return errors.NewInternal(err)
	}

	return nil
}
