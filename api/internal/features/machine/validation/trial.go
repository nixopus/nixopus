package validation

import (
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
)

// Validator handles validation for trial requests.
type Validator struct{}

// NewValidator creates a new Validator instance.
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRequest validates request objects using type switch.
func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *machine_types.ProvisionRequest:
		return v.validateProvisionRequest(*r)
	default:
		return machine_types.ErrInvalidRequestType
	}
}

func (v *Validator) validateProvisionRequest(req machine_types.ProvisionRequest) error {
	return nil
}
