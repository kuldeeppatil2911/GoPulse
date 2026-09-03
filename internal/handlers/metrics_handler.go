package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gopulse/internal/errors"
	"gopulse/internal/models"
	"gopulse/internal/services"
)

type MetricsHandler struct {
	metricsService services.MetricsService
}

func NewMetricsHandler(ms services.MetricsService) *MetricsHandler {
	return &MetricsHandler{metricsService: ms}
}

type MetricPayload struct {
	RequestID  string    `json:"requestId" binding:"required"`
	Method     string    `json:"method" binding:"required"`
	Path       string    `json:"path" binding:"required"`
	StatusCode int       `json:"statusCode" binding:"required"`
	LatencyMs  int       `json:"latencyMs" binding:"required"`
	ClientIP   *string   `json:"clientIp"`
	Timestamp  time.Time `json:"timestamp"`
}

func (h *MetricsHandler) RecordMetric(c *gin.Context) {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		c.Error(errors.NewUnauthorized("Missing X-API-Key header"))
		return
	}

	var payload MetricPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.Error(errors.NewInvalidRequest("Invalid metric payload: " + err.Error()))
		return
	}

	metric := &models.RequestMetric{
		RequestID:  payload.RequestID,
		Method:     payload.Method,
		Path:       payload.Path,
		StatusCode: payload.StatusCode,
		LatencyMs:  payload.LatencyMs,
		ClientIP:   payload.ClientIP,
		Timestamp:  payload.Timestamp,
	}

	if err := h.metricsService.RecordMetric(c.Request.Context(), apiKey, metric); err != nil {
		c.Error(err)
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"success": true})
}
