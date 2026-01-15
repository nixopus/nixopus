package types

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// HealthCheck represents a health check configuration for an application
type HealthCheck struct {
	bun.BaseModel    `bun:"table:health_checks,alias:hc" swaggerignore:"true"`
	ID               uuid.UUID            `json:"id" bun:"id,pk,type:uuid"`
	ApplicationID    uuid.UUID            `json:"application_id" bun:"application_id,notnull,type:uuid"`
	OrganizationID   uuid.UUID            `json:"organization_id" bun:"organization_id,notnull,type:uuid"`
	Enabled          bool                 `json:"enabled" bun:"enabled,notnull,default:true"`
	Endpoint         string               `json:"endpoint" bun:"endpoint,notnull,default:'/'"`
	Method           string               `json:"method" bun:"method,notnull,default:'GET'"`
	ExpectedStatus   []int                `json:"expected_status_codes" bun:"expected_status_codes,array"`
	TimeoutSeconds   int                  `json:"timeout_seconds" bun:"timeout_seconds,notnull,default:30"`
	IntervalSeconds  int                  `json:"interval_seconds" bun:"interval_seconds,notnull,default:60"`
	FailureThreshold int                  `json:"failure_threshold" bun:"failure_threshold,notnull,default:3"`
	SuccessThreshold int                  `json:"success_threshold" bun:"success_threshold,notnull,default:1"`
	Headers          map[string]string    `json:"headers,omitempty" bun:"headers,type:jsonb"`
	Body             string               `json:"body,omitempty" bun:"body"`
	ConsecutiveFails int                  `json:"consecutive_fails" bun:"consecutive_fails,notnull,default:0"`
	LastCheckedAt    *time.Time           `json:"last_checked_at,omitempty" bun:"last_checked_at"`
	RetentionDays    int                  `json:"retention_days" bun:"retention_days,notnull,default:30"`
	CreatedAt        time.Time            `json:"created_at" bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt        time.Time            `json:"updated_at" bun:"updated_at,notnull,default:current_timestamp"`
	Application      *Application         `json:"application,omitempty" bun:"rel:belongs-to,join:application_id=id"`
	Results          []*HealthCheckResult `json:"results,omitempty" bun:"rel:has-many,join:id=health_check_id"`
}

// HealthCheckResult represents a single health check execution result
type HealthCheckResult struct {
	bun.BaseModel  `bun:"table:health_check_results,alias:hcr" swaggerignore:"true"`
	ID             uuid.UUID    `json:"id" bun:"id,pk,type:uuid"`
	HealthCheckID  uuid.UUID    `json:"health_check_id" bun:"health_check_id,notnull,type:uuid"`
	Status         string       `json:"status" bun:"status,notnull"`
	ResponseTimeMs int          `json:"response_time_ms" bun:"response_time_ms"`
	StatusCode     int          `json:"status_code,omitempty" bun:"status_code"`
	ErrorMessage   string       `json:"error_message,omitempty" bun:"error_message"`
	CheckedAt      time.Time    `json:"checked_at" bun:"checked_at,notnull,default:current_timestamp"`
	HealthCheck    *HealthCheck `json:"health_check,omitempty" bun:"rel:belongs-to,join:health_check_id=id"`
}

// HealthCheckStatus represents the current status of a health check
type HealthCheckStatus string

const (
	HealthCheckStatusHealthy   HealthCheckStatus = "healthy"
	HealthCheckStatusUnhealthy HealthCheckStatus = "unhealthy"
	HealthCheckStatusTimeout   HealthCheckStatus = "timeout"
	HealthCheckStatusError     HealthCheckStatus = "error"
	HealthCheckStatusUnknown   HealthCheckStatus = "unknown"
)
