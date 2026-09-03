package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"gopulse/internal/errors"
	"gopulse/internal/handlers"
	"gopulse/internal/middleware"
	"gopulse/internal/models"
)

// Mock MetricsService
type mockMetricsService struct {
	RecordMetricFunc func(ctx context.Context, apiKey string, metric *models.RequestMetric) error
	RecordErrorFunc  func(ctx context.Context, apiKey string, errModel *models.Error) error
}

func (m *mockMetricsService) RecordMetric(ctx context.Context, apiKey string, metric *models.RequestMetric) error {
	if m.RecordMetricFunc != nil {
		return m.RecordMetricFunc(ctx, apiKey, metric)
	}
	return nil
}

func (m *mockMetricsService) RecordError(ctx context.Context, apiKey string, errModel *models.Error) error {
	if m.RecordErrorFunc != nil {
		return m.RecordErrorFunc(ctx, apiKey, errModel)
	}
	return nil
}

func setupRouter(ms *mockMetricsService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.ErrorHandlerMiddleware())
	
	handler := handlers.NewMetricsHandler(ms)
	r.POST("/api/metrics", handler.RecordMetric)
	
	return r
}

func TestRecordMetric_HTTP_Success(t *testing.T) {
	mockSvc := &mockMetricsService{
		RecordMetricFunc: func(ctx context.Context, apiKey string, metric *models.RequestMetric) error {
			if apiKey != "test-api-key" {
				return errors.NewUnauthorized("Invalid API Key")
			}
			return nil
		},
	}

	router := setupRouter(mockSvc)

	payload := handlers.MetricPayload{
		RequestID:  "req-123",
		Method:     "POST",
		Path:       "/api/charge",
		StatusCode: 200,
		LatencyMs:  125,
		Timestamp:  time.Now(),
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/metrics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "test-api-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("Expected status code %d, got %d. Body: %s", http.StatusAccepted, w.Code, w.Body.String())
	}
}

func TestRecordMetric_HTTP_InvalidKey(t *testing.T) {
	mockSvc := &mockMetricsService{
		RecordMetricFunc: func(ctx context.Context, apiKey string, metric *models.RequestMetric) error {
			return errors.NewUnauthorized("Invalid API Key")
		},
	}

	router := setupRouter(mockSvc)

	payload := handlers.MetricPayload{
		RequestID:  "req-123",
		Method:     "POST",
		Path:       "/api/charge",
		StatusCode: 200,
		LatencyMs:  125,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequest(http.MethodPost, "/api/metrics", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", "wrong-key")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("Expected status code %d, got %d. Body: %s", http.StatusUnauthorized, w.Code, w.Body.String())
	}
}
