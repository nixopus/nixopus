package validation

import (
	"encoding/json"
	"io"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateID(id string) error {
	if id == "" {
		return types.ErrMissingDomainID
	}
	if _, err := uuid.Parse(id); err != nil {
		return types.ErrInvalidDomainID
	}
	return nil
}

func (v *Validator) ValidateName(domain string) error {
	if domain == "" {
		return types.ErrMissingDomainName
	}

	return nil
}

func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *types.CreateDomainRequest:
		return v.validateCreateDomainRequest(*r)
	case *types.UpdateDomainRequest:
		return v.validateUpdateDomainRequest(*r)
	case *types.DeleteDomainRequest:
		return v.validateDeleteDomainRequest(*r)
	default:
		return types.ErrInvalidRequestType
	}
}

func (v *Validator) validateCreateDomainRequest(req types.CreateDomainRequest) error {
	if err := v.ValidateName(req.Name); err != nil {
		return err
	}
	return nil
}

func (v *Validator) validateUpdateDomainRequest(req types.UpdateDomainRequest) error {
	if err := v.ValidateName(req.Name); err != nil {
		return err
	}
	return nil
}

func (v *Validator) validateDeleteDomainRequest(req types.DeleteDomainRequest) error {
	if err := v.ValidateID(req.ID); err != nil {
		return err
	}
	return nil
}
