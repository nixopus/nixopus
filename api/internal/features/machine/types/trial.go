package types

import "errors"

type ProvisionRequest struct {
	Image string `json:"image,omitempty"`
}

type ProvisionResponse struct {
	SessionID string `json:"session_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
}

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

type ProvisionTrailResponse struct {
	Status  string             `json:"status"`
	Message string             `json:"message,omitempty"`
	Data    *ProvisionResponse `json:"data,omitempty"`
	Error   string             `json:"error,omitempty"`
}

type TrailStatusEnvelopeResponse struct {
	Status  string          `json:"status"`
	Message string          `json:"message,omitempty"`
	Data    *StatusResponse `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type UpgradeResourcesResponse struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
}

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

type UserProvisionStatus string

const (
	UserProvisionStatusPending      UserProvisionStatus = "pending"
	UserProvisionStatusProvisioning UserProvisionStatus = "provisioning"
	UserProvisionStatusCompleted    UserProvisionStatus = "completed"
	UserProvisionStatusFailed       UserProvisionStatus = "failed"
)

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
