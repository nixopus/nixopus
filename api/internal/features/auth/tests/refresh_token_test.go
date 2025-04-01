package tests

import (
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRefreshToken(t *testing.T) {
	setup := testutils.NewTestSetup()

	registerRequest := types.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
		Type:     "viewer",
	}

	registerResponse, err := setup.AuthService.Register(registerRequest, "app_user")
	assert.NoError(t, err)

	tests := []struct {
		name          string
		request       types.RefreshTokenRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "successful token refresh",
			request: types.RefreshTokenRequest{
				RefreshToken: registerResponse.RefreshToken,
			},
			expectError: false,
		},
		{
			name: "invalid refresh token",
			request: types.RefreshTokenRequest{
				RefreshToken: "invalid_token",
			},
			expectError:   true,
			errorContains: types.ErrRefreshTokenAlreadyRevoked.Error(),
		},
		{
			name: "empty refresh token",
			request: types.RefreshTokenRequest{
				RefreshToken: "",
			},
			expectError:   true,
			errorContains: types.ErrRefreshTokenIsRequired.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := setup.AuthService.RefreshToken(tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, response.AccessToken)
			assert.NotEmpty(t, response.RefreshToken)
			assert.NotEmpty(t, response.User.ID)
			assert.Equal(t, registerRequest.Email, response.User.Email)
		})
	}
}

func TestRefreshTokenWithRevokedToken(t *testing.T) {
	setup := testutils.NewTestSetup()

	registerRequest := types.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
		Type:     "viewer",
	}

	registerResponse, err := setup.AuthService.Register(registerRequest, "app_user")
	assert.NoError(t, err)

	err = setup.AuthService.Logout(registerResponse.RefreshToken)
	assert.NoError(t, err)

	_, err = setup.AuthService.RefreshToken(types.RefreshTokenRequest{
		RefreshToken: registerResponse.RefreshToken,
	})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), types.ErrRefreshTokenAlreadyRevoked.Error())
}
