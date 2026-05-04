package validation

import (
	"fmt"
	"strings"

	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// GithubConnectorRepository defines the interface for the storage dependency
// This makes it easier to mock in tests
type GithubConnectorRepository interface {
	GetAllConnectors(userID string) ([]shared_types.GithubConnector, error)
}

// Validator handles validation logic for github connector
type Validator struct {
	storage GithubConnectorRepository
	Logger  *logger.Logger // optional; nil disables validation logs
}

// NewValidator creates a new validator instance
func NewValidator(storage GithubConnectorRepository) *Validator {
	return NewValidatorWithLogger(storage, nil)
}

// NewValidatorWithLogger is like NewValidator but attaches a logger for Debug detail on rule failures.
func NewValidatorWithLogger(storage GithubConnectorRepository, log *logger.Logger) *Validator {
	return &Validator{
		storage: storage,
		Logger:  log,
	}
}

func (v *Validator) valog(sev logger.Severity, msg, data string) {
	if v == nil || v.Logger == nil {
		return
	}
	v.Logger.Log(sev, msg, data)
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
		v.valog(logger.Debug, "github connector validation: invalid request type", fmt.Sprintf("%T", req))
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
			v.valog(logger.Debug, "github connector validation: create rejected", "reason=missing_slug")
			return types.ErrMissingSlug
		}
		if isEmpty(req.Pem) {
			v.valog(logger.Debug, "github connector validation: create rejected", "reason=missing_pem")
			return types.ErrMissingPem
		}
		if isEmpty(req.ClientID) {
			v.valog(logger.Debug, "github connector validation: create rejected", "reason=missing_client_id")
			return types.ErrMissingClientID
		}
		if isEmpty(req.ClientSecret) {
			v.valog(logger.Debug, "github connector validation: create rejected", "reason=missing_client_secret")
			return types.ErrMissingClientSecret
		}
		if isEmpty(req.WebhookSecret) {
			v.valog(logger.Debug, "github connector validation: create rejected", "reason=missing_webhook_secret")
			return types.ErrMissingWebhookSecret
		}
	}
	// If no credentials provided, validation passes (will use shared config)
	return nil
}

func (v *Validator) validateUpdateGithubConnectorRequest(req types.UpdateGithubConnectorRequest) error {
	if req.InstallationID == "" {
		v.valog(logger.Debug, "github connector validation: update rejected", "reason=missing_installation_id")
		return types.ErrMissingInstallationID
	}

	return nil
}

func (v *Validator) validateDeleteGithubConnectorRequest(req types.DeleteGithubConnectorRequest) error {
	if req.ID == "" {
		v.valog(logger.Debug, "github connector validation: delete rejected", "reason=missing_id")
		return types.ErrMissingID
	}

	return nil
}
