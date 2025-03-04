package validation

import (
	"encoding/json"
	"io"

	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
)

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case types.CreateGithubConnectorRequest:
		return v.validateCreateGithubConnectorRequest(r)
	case types.UpdateGithubConnectorRequest:
		return v.validateUpdateGithubConnectorRequest(r)
	default:
		return types.ErrInvalidRequestType
	}
}

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

func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}
