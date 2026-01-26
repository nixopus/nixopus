package live

import (
	"time"

	"github.com/google/uuid"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// CreateSessionRequest represents the request body for creating a live session
type CreateSessionRequest struct {
	ApplicationID  uuid.UUID              `json:"application_id"`
	UserID         uuid.UUID              `json:"user_id"`
	OrganizationID uuid.UUID              `json:"organization_id"`
	Config         map[string]interface{} `json:"config"`
}

// CreateSessionResponse represents the response from session creation endpoint
type CreateSessionResponse struct {
	SessionID   uuid.UUID `json:"session_id"`
	StagingPath string    `json:"staging_path"`
	Status      string    `json:"status"`
	StartedAt   time.Time `json:"started_at"`
	Domain      string    `json:"domain"`
}

// StatusResponse represents a generic status response
type StatusResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// ApplicationContext holds application information for live deployments
type ApplicationContext struct {
	ApplicationID        uuid.UUID
	UserID               uuid.UUID
	OrganizationID       uuid.UUID
	StagingPath          string
	BasePath             string // Base path for monorepo apps (e.g., "api", "apps/frontend")
	Environment          shared_types.Environment
	Domain               string
	Config               map[string]interface{}
	EnvironmentVariables map[string]string
}
