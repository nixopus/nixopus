package validation

import (
	"github.com/raghavyuva/nixopus-api/internal/features/trail/types"
)

// Validator handles validation for trail requests.
type Validator struct{}

// NewValidator creates a new Validator instance.
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRequest validates request objects using type switch.
//
// Parameters:
//   - req: the request object to validate
//
// Returns:
//   - error: validation error if request is invalid
func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *types.ProvisionRequest:
		return v.validateProvisionRequest(*r)
	default:
		return types.ErrInvalidRequestType
	}
}

// validateProvisionRequest validates a provision request.
// Image is optional, so no validation needed for now.
func (v *Validator) validateProvisionRequest(req types.ProvisionRequest) error {
	return nil
}
