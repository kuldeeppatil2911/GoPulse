package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"gopulse/internal/cache"
	"gopulse/internal/config"
	"gopulse/internal/database"
	"gopulse/internal/graphql"
	"gopulse/internal/handlers"
	"gopulse/internal/logger"
	"gopulse/internal/repositories"
	"gopulse/internal/services"
)

func main() {
	cfg := config.Load()

	if err := logger.InitLogger(cfg.LogLevel, cfg.Environment); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Log.Info("Starting GoPulse API Observability Platform",
		zap.String("environment", cfg.Environment),
		zap.String("port", cfg.Port),
	)

	// Database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Log.Fatal("Database connection failed", zap.Error(err))
	}
	defer db.Close()

	// Cache
	redisCache, err := cache.NewRedisCache(cfg.RedisURL)
	if err != nil {
		logger.Log.Fatal("Redis connection failed", zap.Error(err))
	}

	// Repositories
	serviceRepo := repositories.NewServiceRepository(db)
	endpointRepo := repositories.NewEndpointRepository(db)
	metricsRepo := repositories.NewMetricsRepository(db)
	errorRepo := repositories.NewErrorRepository(db)

	// Services
	metricsService := services.NewMetricsService(serviceRepo, endpointRepo, metricsRepo, errorRepo, redisCache)

	// Handlers
	metricsHandler := handlers.NewMetricsHandler(metricsService)
	serviceHandler := handlers.NewServiceHandler(serviceRepo)

	// Router
	router := handlers.SetupRouter(cfg, metricsHandler, serviceHandler)

	// GraphQL
	gqlResolver := graphql.NewResolver(metricsService)
	gqlServer := handler.NewDefaultServer(graphql.NewExecutableSchema(graphql.Config{Resolvers: gqlResolver}))

	router.POST("/graphql", func(c *gin.Context) {
		gqlServer.ServeHTTP(c.Writer, c.Request)
	})
	router.GET("/playground", func(c *gin.Context) {
		playground.Handler("GraphQL Playground", "/graphql").ServeHTTP(c.Writer, c.Request)
	})

	// Server setup
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Graceful Shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed", zap.Error(err))
		}
	}()

	logger.Log.Info("Server is running", zap.String("port", cfg.Port))

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Log.Info("Server exiting")
}
