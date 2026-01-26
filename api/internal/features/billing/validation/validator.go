package validation

import (
	"github.com/raghavyuva/nixopus-api/internal/features/billing/types"
)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRequest validates request objects using type switch
func (v *Validator) ValidateRequest(req any) error {
	switch r := req.(type) {
	case *types.CreateCheckoutRequest:
		return v.validateCreateCheckoutRequest(*r)
	default:
		return types.ErrInvalidRequestType
	}
}

func (v *Validator) validateCreateCheckoutRequest(req types.CreateCheckoutRequest) error {
	if req.SuccessURL == "" {
		return types.ErrMissingSuccessURL
	}
	if req.CancelURL == "" {
		return types.ErrMissingCancelURL
	}
	return nil
}
