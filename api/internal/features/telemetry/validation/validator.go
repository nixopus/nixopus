package validation

import (
	"regexp"

	"github.com/nixopus/nixopus/api/internal/features/telemetry/types"
)

var semverRegex = regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?$`)

type Validator struct{}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateRequest(req any) error {
	switch r := req.(type) {
	case *types.TrackInstallRequest:
		return v.validateTrackInstall(*r)
	default:
		return types.ErrInvalidRequestType
	}
}

func (v *Validator) validateTrackInstall(req types.TrackInstallRequest) error {
	if !types.AllowedEventTypes[req.EventType] {
		return types.ErrInvalidEventType
	}

	if !types.AllowedOS[req.OS] {
		return types.ErrInvalidOS
	}

	if !types.AllowedArch[req.Arch] {
		return types.ErrInvalidArch
	}

	if !semverRegex.MatchString(req.Version) {
		return types.ErrInvalidVersion
	}

	if req.Duration < 0 || req.Duration > 7200 {
		return types.ErrInvalidDuration
	}

	if len(req.Error) > 200 {
		return types.ErrErrorTooLong
	}

	return nil
}
