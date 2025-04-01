package tests

import (
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestResetPassword(t *testing.T) {
	setup := testutils.NewTestSetup()

	registerRequest := types.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
		Type:     "viewer",
	}

	registerResponse, err := setup.AuthService.Register(registerRequest, "app_user")
	assert.NoError(t, err)

	user, resetToken, err := setup.AuthService.GeneratePasswordResetLink(&registerResponse.User)
	assert.NoError(t, err)
	assert.NotEmpty(t, resetToken)

	tests := []struct {
		name          string
		request       types.ResetPasswordRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "successful password reset",
			request: types.ResetPasswordRequest{
				Password: "newpassword123",
			},
			expectError: false,
		},
		{
			name: "empty password",
			request: types.ResetPasswordRequest{
				Password: "",
			},
			expectError:   true,
			errorContains: types.ErrInvalidResetToken.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := setup.AuthService.ResetPassword(user, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			assert.NoError(t, err)

			response, err := setup.AuthService.Login(user.Email, tt.request.Password)
			assert.NoError(t, err)
			assert.NotEmpty(t, response.AccessToken)
		})
	}
}

func TestResetPasswordWithExpiredToken(t *testing.T) {
	setup := testutils.NewTestSetup()

	registerRequest := types.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
		Type:     "viewer",
	}

	registerResponse, err := setup.AuthService.Register(registerRequest, "app_user")
	assert.NoError(t, err)

	user, resetToken, err := setup.AuthService.GeneratePasswordResetLink(&registerResponse.User)
	assert.NoError(t, err)
	assert.NotEmpty(t, resetToken)

	err = setup.AuthService.ResetPassword(user, types.ResetPasswordRequest{
		Password: "newpassword123",
	})
	assert.NoError(t, err)

	err = setup.AuthService.ResetPassword(user, types.ResetPasswordRequest{
		Password: "newpassword456",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid reset token")
}
