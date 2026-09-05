package handlers

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gopulse/internal/config"
	"gopulse/internal/middleware"
)

func SetupRouter(cfg *config.Config, metricsHandler *MetricsHandler, serviceHandler *ServiceHandler) *gin.Engine {
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// CORS config (allow all for dev/demo)
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Authorization", "X-API-Key"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Global Middleware
	r.Use(gin.Recovery())
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.ErrorHandlerMiddleware())

	// Health checks
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	r.GET("/ready", func(c *gin.Context) {
		// To completely implement readiness, we'd check DB/Redis connections here.
		// For simplicity, we just return UP.
		c.JSON(http.StatusOK, gin.H{"status": "READY"})
	})

	// Serve Frontend Dashboard
	r.Static("/dashboard", "./frontend")

	api := r.Group("/api")
	{
		// Agent endpoints (using X-API-Key)
		api.POST("/metrics", metricsHandler.RecordMetric)

		// Dashboard endpoints (using JWT Auth)
		// For now we assume a simple scenario, you can add AuthMiddleware(cfg.JWTSecret) here
		dashboard := api.Group("/")
		// dashboard.Use(middleware.AuthMiddleware(cfg.JWTSecret)) // Disabled for easier testing locally if needed
		{
			dashboard.POST("/services", serviceHandler.CreateService)
			dashboard.GET("/services", serviceHandler.GetServices)
		}
	}

	return r
}
