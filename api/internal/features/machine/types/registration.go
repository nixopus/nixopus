package types

import "errors"

type CreateMachineRequest struct {
	Name string `json:"name" validate:"required"`
	Host string `json:"host" validate:"required"`
	Port int    `json:"port"`
	User string `json:"user"`
}

type CreateMachineResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	PublicKey string `json:"public_key"`
}

type VerifyMachineResponse struct {
	Status string `json:"status"`
}

type ProvisionMachineRequest struct {
	Name string `json:"name"`
}

type ProvisionMachineResponse struct {
	ProvisionID string `json:"provision_id"`
	Step        string `json:"step"`
	Status      string `json:"status"`
}

type ProvisionStatusResponse struct {
	ProvisionID string  `json:"provision_id"`
	Step        string  `json:"step"`
	Status      string  `json:"status"`
	Error       *string `json:"error"`
}

type DeleteMachineResponse struct {
	Status string `json:"status"`
}

type SSHStatusResponse struct {
	IsActive   bool   `json:"is_active"`
	LastUsedAt string `json:"last_used_at,omitempty"`
}

var (
	ErrFeatureDisabled     = errors.New("feature is disabled")
	ErrMachineLimitReached = errors.New("machine limit reached")
	ErrDuplicateHost       = errors.New("duplicate host and port for this organization")
	ErrMachineHasApps      = errors.New("machine has active deployments")
	ErrInsufficientCredits = errors.New("insufficient credits")
)
