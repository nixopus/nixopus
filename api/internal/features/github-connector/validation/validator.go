package validation

import (
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
// - types.CreateIssueRequest
// - types.UpdateIssueRequest
// - types.CommentOnIssueRequest
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
	case *types.CreateIssueRequest:
		return v.validateCreateIssueRequest(*r)
	case *types.UpdateIssueRequest:
		return v.validateUpdateIssueRequest(*r)
	case *types.CommentOnIssueRequest:
		return v.validateCommentOnIssueRequest(*r)
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
// If any of these fields are empty, an error specific to the missing field
// is returned. Otherwise, nil is returned.
func (v *Validator) validateCreateGithubConnectorRequest(req types.CreateGithubConnectorRequest) error {
	if req.Slug == "" {
		return types.ErrMissingSlug
	}
	if req.Pem == "" {
		return types.ErrMissingPem
	}
	if req.ClientID == "" {
		return types.ErrMissingClientID
	}
	if req.ClientSecret == "" {
		return types.ErrMissingClientSecret
	}
	if req.WebhookSecret == "" {
		return types.ErrMissingWebhookSecret
	}
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

func (v *Validator) validateCreateIssueRequest(req types.CreateIssueRequest) error {
	if req.RepositoryOwner == "" {
		return types.ErrMissingRepositoryOwner
	}
	if req.RepositoryName == "" {
		return types.ErrMissingRepositoryName
	}
	if req.Title == "" {
		return types.ErrMissingTitle
	}
	return nil
}

func (v *Validator) validateUpdateIssueRequest(req types.UpdateIssueRequest) error {
	if req.RepositoryOwner == "" {
		return types.ErrMissingRepositoryOwner
	}
	if req.RepositoryName == "" {
		return types.ErrMissingRepositoryName
	}
	if req.IssueNumber == 0 {
		return types.ErrMissingIssueNumber
	}
	return nil
}

func (v *Validator) validateCommentOnIssueRequest(req types.CommentOnIssueRequest) error {
	if req.RepositoryOwner == "" {
		return types.ErrMissingRepositoryOwner
	}
	if req.RepositoryName == "" {
		return types.ErrMissingRepositoryName
	}
	if req.IssueNumber == 0 {
		return types.ErrMissingIssueNumber
	}
	if req.Body == "" {
		return types.ErrMissingBody
	}
	return nil
}
