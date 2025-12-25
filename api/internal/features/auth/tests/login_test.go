package tests

import (
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestLogin(t *testing.T) {
	setup := testutils.NewTestSetup()

	registerRequest := types.RegisterRequest{
		Email:    "test@example.com",
		Password: "password123",
		Username: "testuser",
		Type:     "viewer",
	}

	_, err := setup.AuthService.Register(registerRequest, "app_user")
	assert.NoError(t, err)

	tests := []struct {
		name          string
		email         string
		password      string
		expectError   bool
		errorContains string
	}{
		{
			name:        "successful login",
			email:       "test@example.com",
			password:    "password123",
			expectError: false,
		},
		{
			name:          "wrong password",
			email:         "test@example.com",
			password:      "wrongpassword",
			expectError:   true,
			errorContains: "invalid password",
		},
		{
			name:          "non-existent user",
			email:         "nonexistent@example.com",
			password:      "password123",
			expectError:   true,
			errorContains: "user not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := setup.AuthService.Login(tt.email, tt.password)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, response.AccessToken)
			assert.NotEmpty(t, response.RefreshToken)
			assert.NotEmpty(t, response.User.ID)
			assert.Equal(t, tt.email, response.User.Email)
		})
	}
}
