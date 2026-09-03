package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopulse/internal/errors"
	"gopulse/internal/models"
	"gopulse/internal/repositories"
)

type ServiceHandler struct {
	serviceRepo repositories.ServiceRepository
}

func NewServiceHandler(repo repositories.ServiceRepository) *ServiceHandler {
	return &ServiceHandler{serviceRepo: repo}
}

type CreateServicePayload struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

func (h *ServiceHandler) CreateService(c *gin.Context) {
	var payload CreateServicePayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.Error(errors.NewInvalidRequest("Invalid payload: " + err.Error()))
		return
	}

	service := &models.Service{
		Name:        payload.Name,
		Description: payload.Description,
		APIKey:      uuid.NewString(), // Generate a simple API key
	}

	if err := h.serviceRepo.Create(c.Request.Context(), service); err != nil {
		c.Error(errors.NewInternal(err))
		return
	}

	// For the response, we want to return the API Key so the user can save it
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": gin.H{
			"id":          service.ID,
			"name":        service.Name,
			"description": service.Description,
			"apiKey":      service.APIKey,
			"createdAt":   service.CreatedAt,
		},
	})
}

func (h *ServiceHandler) GetServices(c *gin.Context) {
	services, err := h.serviceRepo.List(c.Request.Context(), 100, 0)
	if err != nil {
		c.Error(errors.NewInternal(err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    services,
	})
}
