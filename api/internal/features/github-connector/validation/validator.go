package validation

import (
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GithubConnectorRepository defines the interface for the storage dependency
// This makes it easier to mock in tests
type GithubConnectorRepository interface {
	GetAllConnectors(userID string) ([]shared_types.GithubConnector, error)
}

// Validator handles validation logic for github connector
type Validator struct {
	storage GithubConnectorRepository
}

// NewValidator creates a new validator instance
func NewValidator(storage GithubConnectorRepository) *Validator {
	return &Validator{
		storage: storage,
	}
}

// ValidateRequest validates a request object against a set of predefined rules.
// It returns an error if the request object is invalid.
//
// The supported request types are:
// - types.CreateGithubConnectorRequest
// - types.UpdateGithubConnectorRequest
// - types.DeleteGithubConnectorRequest
//
// If the request object is not of one of the above types, it returns
// types.ErrInvalidRequestType.
func (v *Validator) ValidateRequest(req any) error {
	switch r := req.(type) {
	case *types.CreateGithubConnectorRequest:
		return v.validateCreateGithubConnectorRequest(*r)
	case *types.UpdateGithubConnectorRequest:
		return v.validateUpdateGithubConnectorRequest(*r)
	case *types.DeleteGithubConnectorRequest:
		return v.validateDeleteGithubConnectorRequest(*r)
	default:
		return types.ErrInvalidRequestType
	}
}

// validateCreateGithubConnectorRequest validates a CreateGithubConnectorRequest.
//
// It checks the following fields for emptiness:
//
//   - Slug
//   - Pem
//   - ClientID
//   - ClientSecret
//   - WebhookSecret
//
// If credentials are provided (at least one non-empty field), all must be provided.
// If all are empty, validation passes (will use shared config).
func (v *Validator) validateCreateGithubConnectorRequest(req types.CreateGithubConnectorRequest) error {
	// Helper to check if string is truly empty (empty or whitespace only)
	isEmpty := func(s string) bool {
		return s == "" || strings.TrimSpace(s) == ""
	}

	// Check if any credential is provided (non-empty)
	hasCredentials := !isEmpty(req.Slug) || !isEmpty(req.Pem) || !isEmpty(req.ClientID) ||
		!isEmpty(req.ClientSecret) || !isEmpty(req.WebhookSecret) || !isEmpty(req.AppID)

	// If any credential is provided, all must be provided (backward compatibility)
	if hasCredentials {
		if isEmpty(req.Slug) {
			return types.ErrMissingSlug
		}
		if isEmpty(req.Pem) {
			return types.ErrMissingPem
		}
		if isEmpty(req.ClientID) {
			return types.ErrMissingClientID
		}
		if isEmpty(req.ClientSecret) {
			return types.ErrMissingClientSecret
		}
		if isEmpty(req.WebhookSecret) {
			return types.ErrMissingWebhookSecret
		}
	}
	// If no credentials provided, validation passes (will use shared config)
	return nil
}

func (v *Validator) validateUpdateGithubConnectorRequest(req types.UpdateGithubConnectorRequest) error {
	if req.InstallationID == "" {
		return types.ErrMissingInstallationID
	}

	return nil
}

func (v *Validator) validateDeleteGithubConnectorRequest(req types.DeleteGithubConnectorRequest) error {
	if req.ID == "" {
		return types.ErrMissingID
	}

	return nil
}
