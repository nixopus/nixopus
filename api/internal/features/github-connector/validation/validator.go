package validation

import (
	"encoding/json"
	"io"

	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type Validator struct {
	storage storage.GithubConnectorRepository
}

func NewValidator(storage storage.GithubConnectorRepository) *Validator {
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
//
// If the request object is not of one of the above types, it returns
// types.ErrInvalidRequestType.
func (v *Validator) ValidateRequest(req interface{}, user *shared_types.User) error {
	switch r := req.(type) {
	case *types.CreateGithubConnectorRequest:
		return v.validateCreateGithubConnectorRequest(*r)
	case *types.UpdateGithubConnectorRequest:
		return v.validateUpdateGithubConnectorRequest(*r, *user)
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

// validateUpdateGithubConnectorRequest validates the update GitHub connector request.
//
// The method first checks if the InstallationID is empty. If so, it returns an error.
//
// Then, it retrieves all GitHub connectors associated with the provided userID.
// If there are no connectors or if the retrieval fails, it returns an error.
//
// Otherwise, it checks if the first connector's UserID matches the provided userID.
// If not, it returns a permission denied error.
//
// Finally, it returns nil if the validation is successful.
func (v *Validator) validateUpdateGithubConnectorRequest(req types.UpdateGithubConnectorRequest, user shared_types.User) error {
	if req.InstallationID == "" {
		return types.ErrMissingInstallationID
	}

	connectors, err := v.storage.GetAllConnectors(user.ID.String())

	if err != nil {
		return err
	}
	if len(connectors) == 0 {
		return types.ErrNoConnectors
	}

	if string(connectors[0].UserID.String()) != user.ID.String() || user.Type != "admin" {
		return types.ErrPermissionDenied
	}

	return nil
}

// ParseRequestBody parses a request body into the provided interface.
//
// The method decodes the body of the request using the provided io.ReadCloser,
// and stores the result in the provided interface.
//
// If the decoding fails, it returns an error.
func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}
