package validation

import (
	"encoding/json"
	"io"

	"github.com/raghavyuva/nixopus-api/internal/features/role/types"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) validateCreateRoleRequest(role types.CreateRoleRequest) error {
	if role.Name == "" {
		return types.ErrRoleNameRequired
	}
	return nil
}

func (v *Validator) validateGetRoleRequest(id string) error {
	if id == "" {
		return types.ErrRoleIDRequired
	}
	return nil
}

func (v *Validator) validateUpdateRoleRequest(role types.UpdateRoleRequest) error {
	if role.ID == "" {
		return types.ErrRoleIDRequired
	}
	if role.Name == "" && role.Description == "" {
		return types.ErrRoleEmptyFields
	}
	return nil
}

func (v *Validator) ParseRequestBody(req interface{}, body io.ReadCloser, decoded interface{}) error {
	return json.NewDecoder(body).Decode(decoded)
}

func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case types.CreateRoleRequest:
		return v.validateCreateRoleRequest(r)
	case types.GetRoleRequest:
		return v.validateGetRoleRequest(r.ID)
	case types.UpdateRoleRequest:
		return v.validateUpdateRoleRequest(r)
	default:
		return types.ErrInvalidRequestType
	}
}
