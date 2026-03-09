package types

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// ProvisionRequest represents a trail provisioning request.
type ProvisionRequest struct {
	Image string `json:"image,omitempty"`
}

// ProvisionResponse represents the response after initiating provisioning.
type ProvisionResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

// StatusResponse represents the current status of a trail provision.
type StatusResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Step      string `json:"step,omitempty"`
	Progress  int    `json:"progress"`
	Message   string `json:"message"`
	Subdomain string `json:"subdomain,omitempty"`
	TrailURL  string `json:"trail_url,omitempty"`
}

type UpgradeResourcesRequest struct {
	UserID    string `json:"user_id"`
	OrgID     string `json:"org_id"`
	VcpuCount int    `json:"vcpu_count"`
	MemoryMB  int    `json:"memory_mb"`
}

// ProvisionTrailResponse is a typed response for provisioning requests.
type ProvisionTrailResponse struct {
	Status  string             `json:"status"`
	Message string             `json:"message,omitempty"`
	Data    *ProvisionResponse `json:"data,omitempty"`
	Error   string             `json:"error,omitempty"`
}

// TrailStatusEnvelopeResponse is a typed response for trail status retrieval.
type TrailStatusEnvelopeResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message,omitempty"`
	Data    *StatusResponse `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// UpgradeResourcesResponse is a typed message response for resource upgrades.
type UpgradeResourcesResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ProvisionPayload represents the payload sent to the Redis queue for async processing.
type ProvisionPayload struct {
	SessionID          string `json:"session_id"`
	Subdomain          string `json:"subdomain"`
	ContainerName      string `json:"container_name"`
	Image              string `json:"image"`
	UserID             string `json:"user_id"`
	OrgID              string `json:"org_id"`
	ProvisionDetailsID string `json:"provision_details_id"`
	ServerID           string `json:"server_id,omitempty"`
}

// ProvisionStep represents the current step in the provisioning process.
type ProvisionStep string

const (
	ProvisionStepInitializing       ProvisionStep = "INITIALIZING"
	ProvisionStepCreatingContainer  ProvisionStep = "CREATING_CONTAINER"
	ProvisionStepSetupNetworking    ProvisionStep = "SETUP_NETWORKING"
	ProvisionStepInstallingDeps     ProvisionStep = "INSTALLING_DEPENDENCIES"
	ProvisionStepConfiguringSSH     ProvisionStep = "CONFIGURING_SSH"
	ProvisionStepSetupSSHForwarding ProvisionStep = "SETUP_SSH_FORWARDING"
	ProvisionStepVerifyingSSH       ProvisionStep = "VERIFYING_SSH"
	ProvisionStepCompleted          ProvisionStep = "COMPLETED"
)

// UserProvisionStatus represents the overall status of a user's provision.
type UserProvisionStatus string

const (
	UserProvisionStatusNotStarted   UserProvisionStatus = "NOT_STARTED"
	UserProvisionStatusProvisioning UserProvisionStatus = "PROVISIONING"
	UserProvisionStatusActive       UserProvisionStatus = "ACTIVE"
	UserProvisionStatusFailed       UserProvisionStatus = "FAILED"
)

// UserProvisionDetails represents the database model for provision details.
type UserProvisionDetails struct {
	bun.BaseModel `bun:"table:user_provision_details,alias:upd" swaggerignore:"true"`

	ID               uuid.UUID      `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	UserID           uuid.UUID      `bun:"user_id,type:uuid,notnull" json:"user_id"`
	OrganizationID   uuid.UUID      `bun:"organization_id,type:uuid,notnull" json:"organization_id"`
	ServerID         *uuid.UUID     `bun:"server_id,type:uuid" json:"server_id,omitempty"`
	LXDContainerName *string        `bun:"lxd_container_name" json:"lxd_container_name,omitempty"`
	Subdomain        *string        `bun:"subdomain" json:"subdomain,omitempty"`
	Domain           *string        `bun:"domain" json:"domain,omitempty"`
	Step             *ProvisionStep `bun:"step" json:"step,omitempty"`
	Error            *string        `bun:"error" json:"error,omitempty"`
	CreatedAt        time.Time      `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt        time.Time      `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// Domain errors for trail provisioning.
var (
	ErrImageNotAllowed       = errors.New("requested image is not allowed")
	ErrActiveProvisionExists = errors.New("you already have an active trail provision")
	ErrSystemAtCapacity      = errors.New("system is at capacity. please try again later")
	ErrProvisionNotFound     = errors.New("provision not found")
	ErrInvalidSessionID      = errors.New("invalid session ID format")
	ErrInvalidRequestType    = errors.New("invalid request type")
	ErrDatabaseNotAvailable  = errors.New("database not available")
	ErrOrganizationRequired  = errors.New("organization context required")
	ErrInvalidOrganizationID = errors.New("invalid organization ID")
	ErrFailedToEnqueueTask   = errors.New("failed to queue provisioning task")
)
