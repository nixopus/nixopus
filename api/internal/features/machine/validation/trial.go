package validation

import (
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
)

// Validator handles validation for trial requests.
type Validator struct {
	Logger *logger.Logger // optional; nil disables validation logs
}

// NewValidator creates a new Validator instance.
func NewValidator() *Validator {
	return NewValidatorWithLogger(nil)
}

// NewValidatorWithLogger creates a Validator with optional structured logging.
func NewValidatorWithLogger(l *logger.Logger) *Validator {
	return &Validator{Logger: l}
}

func (v *Validator) valog(sev logger.Severity, msg, data string) {
	if v.Logger == nil {
		return
	}
	v.Logger.Log(sev, msg, data)
}

// ValidateRequest validates request objects using type switch.
func (v *Validator) ValidateRequest(req interface{}) error {
	switch r := req.(type) {
	case *machine_types.ProvisionRequest:
		return v.validateProvisionRequest(*r)
	default:
		v.valog(logger.Debug, fmt.Sprintf("machine trial validation: invalid type %T", req), "")
		return machine_types.ErrInvalidRequestType
	}
}

func (v *Validator) validateProvisionRequest(req machine_types.ProvisionRequest) error {
	return nil
}
