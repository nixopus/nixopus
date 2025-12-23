package tests

import (
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	setup := testutils.NewTestSetup()

	tests := []struct {
		name          string
		request       types.RegisterRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "successful registration",
			request: types.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Username: "testuser",
				Type:     "viewer",
			},
			expectError: false,
		},
		{
			name: "duplicate email",
			request: types.RegisterRequest{
				Email:    "test@example.com",
				Password: "password123",
				Username: "testuser2",
				Type:     "viewer",
			},
			expectError:   true,
			errorContains: "user with email already exists",
		},
		{
			name: "invalid user type",
			request: types.RegisterRequest{
				Email:    "test2@example.com",
				Password: "password123",
				Username: "testuser3",
				Type:     "invalid_type",
			},
			expectError:   true,
			errorContains: "invalid user type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := setup.AuthService.Register(tt.request, "admin")

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				return
			}

			assert.NoError(t, err)
			assert.NotEmpty(t, response.AccessToken)
			assert.NotEmpty(t, response.RefreshToken)
			assert.NotEmpty(t, response.User.ID)
			assert.Equal(t, tt.request.Email, response.User.Email)
			assert.Equal(t, tt.request.Username, response.User.Username)
		})
	}
}

func TestRegisterWithOrganization(t *testing.T) {
	setup := testutils.NewTestSetup()

	request := types.RegisterRequest{
		Email:    "orgtest@example.com",
		Password: "password123",
		Username: "orguser",
		Type:     "admin",
	}

	response, err := setup.AuthService.Register(request, "admin")
	assert.NoError(t, err)

	assert.NotEmpty(t, response.User.ID)
}
