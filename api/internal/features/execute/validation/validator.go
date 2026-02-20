package validation

import (
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/execute/types"
)

// Validator handles validation for execute requests.
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
	case *types.ExecuteRequest:
		return v.validateExecuteRequest(*r)
	default:
		return types.ErrInvalidRequestType
	}
}

// validateExecuteRequest validates an execute request.
func (v *Validator) validateExecuteRequest(req types.ExecuteRequest) error {
	if req.Command == "" {
		return types.ErrCommandRequired
	}

	baseCommand := strings.TrimSpace(req.Command)
	if !types.AllowedCommands[baseCommand] {
		return types.ErrCommandNotAllowed
	}

	return nil
}
