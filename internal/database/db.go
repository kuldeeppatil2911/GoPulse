package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"gopulse/internal/logger"
)

func Connect(databaseURL string) (*sqlx.DB, error) {
	logger.Log.Info("Connecting to PostgreSQL", zap.String("url", redactURL(databaseURL)))

	db, err := sqlx.Connect("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Log.Info("Successfully connected to PostgreSQL")
	return db, nil
}

// redactURL hides the password for logging
func redactURL(url string) string {
	// Simple redaction logic could go here, for now return "redacted" if needed,
	// or assume the caller doesn't pass a plain password string.
	// We'll just return a placeholder to avoid leaking secrets.
	return "postgres://[redacted]@[host]/[dbname]?sslmode=disable"
}
