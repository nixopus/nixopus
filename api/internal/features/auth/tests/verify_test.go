package tests

import (
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestVerifyToken(t *testing.T) {
	setup := testutils.NewTestSetup()

	registerRequest := types.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
		Type:     "viewer",
	}

	registerResponse, err := setup.AuthService.Register(registerRequest, "app_user")
	assert.NoError(t, err)

	verificationToken, err := setup.AuthService.GenerateVerificationToken(registerResponse.User.ID.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, verificationToken)

	tests := []struct {
		name          string
		token         string
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful verification",
			token:       verificationToken,
			expectError: false,
		},
		{
			name:          "invalid token",
			token:         "invalid_token",
			expectError:   true,
			errorContains: "verification token is already used",
		},
		{
			name:          "empty token",
			token:         "",
			expectError:   true,
			errorContains: "verification token is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userID, err := setup.AuthService.VerifyToken(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, registerResponse.User.ID.String(), userID)

			user, err := setup.AuthService.GetUserByID(userID)
			assert.NoError(t, err)
			assert.True(t, user.IsVerified)
		})
	}
}

func TestVerifyTokenWithExpiredToken(t *testing.T) {
	setup := testutils.NewTestSetup()

	registerRequest := types.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
		Type:     "viewer",
	}

	registerResponse, err := setup.AuthService.Register(registerRequest, "app_user")
	assert.NoError(t, err)

	verificationToken, err := setup.AuthService.GenerateVerificationToken(registerResponse.User.ID.String())
	assert.NoError(t, err)
	assert.NotEmpty(t, verificationToken)

	userID, err := setup.AuthService.VerifyToken(verificationToken)
	assert.NoError(t, err)
	assert.Equal(t, registerResponse.User.ID.String(), userID)

	_, err = setup.AuthService.VerifyToken(verificationToken)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification token is already used")
}
