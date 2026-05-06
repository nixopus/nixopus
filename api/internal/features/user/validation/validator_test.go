package validation_test

import (
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/user/types"
	"github.com/nixopus/nixopus/api/internal/features/user/validation"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewValidator(t *testing.T) {
	v := validation.NewValidator()
	require.NotNil(t, v)
}

func TestNewValidatorWithLogger(t *testing.T) {
	l := logger.NewLogger()
	v := validation.NewValidatorWithLogger(&l)
	require.NotNil(t, v)
	assert.NotNil(t, v.Logger)
}

func TestValidateRequest_WithLogger_InvalidType(t *testing.T) {
	l := logger.NewLogger()
	v := validation.NewValidatorWithLogger(&l)
	user := shared_types.User{}
	err := v.ValidateRequest(struct{}{}, user)
	assert.ErrorIs(t, err, types.ErrInvalidRequestType)
}

func TestValidateRequest(t *testing.T) {
	v := &validation.Validator{}
	user := shared_types.User{}

	tests := []struct {
		name    string
		req     interface{}
		wantErr error
	}{
		{
			name:    "Valid UpdateUserNameRequest",
			req:     &types.UpdateUserNameRequest{Name: "newusername"},
			wantErr: nil,
		},
		{
			name:    "Valid UpdateAvatarRequest",
			req:     &types.UpdateAvatarRequest{AvatarData: "data:image/png;base64,abc"},
			wantErr: nil,
		},
		{
			name:    "Invalid request type",
			req:     struct{}{},
			wantErr: types.ErrInvalidRequestType,
		},
		{
			name:    "Nil request",
			req:     nil,
			wantErr: types.ErrInvalidRequestType,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateRequest(tt.req, user)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUpdateUserNameRequest(t *testing.T) {
	v := &validation.Validator{}
	user := shared_types.User{Username: "currentuser"}

	tests := []struct {
		name    string
		req     types.UpdateUserNameRequest
		wantErr error
	}{
		{
			name:    "Empty username",
			req:     types.UpdateUserNameRequest{Name: ""},
			wantErr: types.ErrUserNameIsEmpty,
		},
		{
			name:    "Same username as current user",
			req:     types.UpdateUserNameRequest{Name: "currentuser"},
			wantErr: types.ErrSameUserName,
		},
		{
			name:    "Username too long",
			req:     types.UpdateUserNameRequest{Name: strings.Repeat("a", 51)},
			wantErr: types.ErrUserNameTooLong,
		},
		{
			name:    "Username contains spaces",
			req:     types.UpdateUserNameRequest{Name: "username with spaces"},
			wantErr: types.ErrUserNameContainsSpaces,
		},
		{
			name:    "Username too short",
			req:     types.UpdateUserNameRequest{Name: "ab"},
			wantErr: types.ErrUsernameTooShort,
		},
		{
			name:    "Valid username",
			req:     types.UpdateUserNameRequest{Name: "newusername"},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateUpdateUserNameRequest(tt.req, user)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("validateUpdateUserNameRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateUpdateAvatarRequest(t *testing.T) {
	v := &validation.Validator{}

	tests := []struct {
		name    string
		data    string
		wantErr error
	}{
		{
			name:    "Empty avatar data",
			data:    "",
			wantErr: types.ErrInvalidAvatarData,
		},
		{
			name:    "Missing data:image/ prefix",
			data:    "base64,abc",
			wantErr: types.ErrInvalidAvatarData,
		},
		{
			name:    "Missing ;base64, separator",
			data:    "data:image/png",
			wantErr: types.ErrInvalidAvatarData,
		},
		{
			name:    "Unsupported image type",
			data:    "data:image/bmp;base64,abc",
			wantErr: types.ErrUnsupportedImageFormat,
		},
		{
			name:    "Valid png",
			data:    "data:image/png;base64,abc",
			wantErr: nil,
		},
		{
			name:    "Valid jpeg",
			data:    "data:image/jpeg;base64,abc",
			wantErr: nil,
		},
		{
			name:    "Valid jpg",
			data:    "data:image/jpg;base64,abc",
			wantErr: nil,
		},
		{
			name:    "Valid gif",
			data:    "data:image/gif;base64,abc",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := v.ValidateUpdateAvatarRequest(types.UpdateAvatarRequest{AvatarData: tt.data})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ValidateUpdateAvatarRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestParseRequestBody(t *testing.T) {
	v := validation.NewValidator()

	t.Run("valid JSON decodes successfully", func(t *testing.T) {
		body := io.NopCloser(strings.NewReader(`{"name":"alice"}`))
		var req types.UpdateUserNameRequest
		err := v.ParseRequestBody(nil, body, &req)
		require.NoError(t, err)
		assert.Equal(t, "alice", req.Name)
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		body := io.NopCloser(strings.NewReader(`not-json`))
		var req types.UpdateUserNameRequest
		err := v.ParseRequestBody(nil, body, &req)
		assert.Error(t, err)
	})
}
