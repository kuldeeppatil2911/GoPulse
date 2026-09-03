package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID `db:"id" json:"id"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time `db:"updated_at" json:"updatedAt"`
}

type Service struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	Description *string   `db:"description" json:"description"`
	APIKey      string    `db:"api_key" json:"-"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}

type Endpoint struct {
	ID        uuid.UUID `db:"id" json:"id"`
	ServiceID uuid.UUID `db:"service_id" json:"serviceId"`
	Path      string    `db:"path" json:"path"`
	Method    string    `db:"method" json:"method"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}

type RequestMetric struct {
	ID         uuid.UUID  `db:"id" json:"id"`
	ServiceID  uuid.UUID  `db:"service_id" json:"serviceId"`
	EndpointID *uuid.UUID `db:"endpoint_id" json:"endpointId,omitempty"`
	RequestID  string     `db:"request_id" json:"requestId"`
	Method     string     `db:"method" json:"method"`
	Path       string     `db:"path" json:"path"`
	StatusCode int        `db:"status_code" json:"statusCode"`
	LatencyMs  int        `db:"latency_ms" json:"latencyMs"`
	ClientIP   *string    `db:"client_ip" json:"clientIp,omitempty"`
	Timestamp  time.Time  `db:"timestamp" json:"timestamp"`
}

type Error struct {
	ID           uuid.UUID `db:"id" json:"id"`
	ServiceID    uuid.UUID `db:"service_id" json:"serviceId"`
	RequestID    *string   `db:"request_id" json:"requestId,omitempty"`
	ErrorCode    string    `db:"error_code" json:"errorCode"`
	ErrorMessage string    `db:"error_message" json:"errorMessage"`
	StackTrace   *string   `db:"stack_trace" json:"stackTrace,omitempty"`
	Timestamp    time.Time `db:"timestamp" json:"timestamp"`
}
