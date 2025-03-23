package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/stretchr/testify/assert"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func TestGeneratePasswordResetLink(t *testing.T) {
	mockStorage := NewMockAuthStorage()
	mockLogger := logger.NewLogger()
	authService := service.NewAuthService(mockStorage, mockLogger, nil, nil, nil, context.Background())

	tests := []struct {
		name        string
		user        *shared_types.User
		tokenErr    error
		updateErr   error
		expectedErr error
		setupMocks  bool
	}{
		{
			name:        "valid user",
			user:        &shared_types.User{Email: "test@example.com"},
			tokenErr:    nil,
			updateErr:   nil,
			expectedErr: nil,
			setupMocks:  true,
		},
		{
			name:        "invalid user (nil)",
			user:        nil,
			tokenErr:    nil,
			updateErr:   nil,
			expectedErr: types.ErrInvalidUser,
			setupMocks:  false,
		},
		{
			name:        "user update error",
			user:        &shared_types.User{Email: "test@example.com"},
			tokenErr:    nil,
			updateErr:   errors.New("user update error"),
			expectedErr: types.ErrFailedToUpdateUser,
			setupMocks:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.setupMocks {
				mockStorage.ExpectedCalls = nil
				mockStorage.On("UpdateUser", test.user).Return(test.updateErr)
			}
			err := authService.GeneratePasswordResetLink(test.user)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}
