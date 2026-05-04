package validation_test

import (
	"strings"
	"testing"

	"github.com/nixopus/nixopus/api/internal/features/telemetry/types"
	"github.com/nixopus/nixopus/api/internal/features/telemetry/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	v := validation.NewValidator()
	require.NotNil(t, v)
}

func TestValidateRequest_unknownType(t *testing.T) {
	v := validation.NewValidator()
	err := v.ValidateRequest("not a known type")
	assert.Equal(t, types.ErrInvalidRequestType, err)
}

func validRequest() *types.TrackInstallRequest {
	return &types.TrackInstallRequest{
		EventType: "install_success",
		OS:        "ubuntu",
		Arch:      "amd64",
		Version:   "1.2.3",
		Duration:  60,
	}
}

func TestValidateRequest_valid(t *testing.T) {
	v := validation.NewValidator()
	require.NoError(t, v.ValidateRequest(validRequest()))
}

func TestValidateRequest_invalidEventType(t *testing.T) {
	v := validation.NewValidator()
	req := validRequest()
	req.EventType = "bad_event"
	assert.Equal(t, types.ErrInvalidEventType, v.ValidateRequest(req))
}

func TestValidateRequest_invalidOS(t *testing.T) {
	v := validation.NewValidator()
	req := validRequest()
	req.OS = "windows"
	assert.Equal(t, types.ErrInvalidOS, v.ValidateRequest(req))
}

func TestValidateRequest_invalidArch(t *testing.T) {
	v := validation.NewValidator()
	req := validRequest()
	req.Arch = "x86"
	assert.Equal(t, types.ErrInvalidArch, v.ValidateRequest(req))
}

func TestValidateRequest_invalidVersion(t *testing.T) {
	v := validation.NewValidator()
	req := validRequest()
	req.Version = "not-a-version"
	assert.Equal(t, types.ErrInvalidVersion, v.ValidateRequest(req))
}

func TestValidateRequest_versionWithPreRelease(t *testing.T) {
	v := validation.NewValidator()
	req := validRequest()
	req.Version = "1.0.0-beta.1"
	require.NoError(t, v.ValidateRequest(req))
}

func TestValidateRequest_durationNegative(t *testing.T) {
	v := validation.NewValidator()
	req := validRequest()
	req.Duration = -1
	assert.Equal(t, types.ErrInvalidDuration, v.ValidateRequest(req))
}

func TestValidateRequest_durationTooLarge(t *testing.T) {
	v := validation.NewValidator()
	req := validRequest()
	req.Duration = 7201
	assert.Equal(t, types.ErrInvalidDuration, v.ValidateRequest(req))
}

func TestValidateRequest_durationBoundary(t *testing.T) {
	v := validation.NewValidator()

	req := validRequest()
	req.Duration = 0
	require.NoError(t, v.ValidateRequest(req))

	req.Duration = 7200
	require.NoError(t, v.ValidateRequest(req))
}

func TestValidateRequest_errorTooLong(t *testing.T) {
	v := validation.NewValidator()
	req := validRequest()
	req.Error = strings.Repeat("x", 201)
	assert.Equal(t, types.ErrErrorTooLong, v.ValidateRequest(req))
}

func TestValidateRequest_errorAtMaxLength(t *testing.T) {
	v := validation.NewValidator()
	req := validRequest()
	req.Error = strings.Repeat("x", 200)
	require.NoError(t, v.ValidateRequest(req))
}
