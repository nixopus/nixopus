package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func TestResetPasswordBasic(t *testing.T) {
	mockStorage := NewMockAuthStorage()
	mockLogger := logger.NewLogger()
	authService := service.NewAuthService(mockStorage, mockLogger, nil, nil, nil, context.Background())

	tests := []struct {
		name           string
		setupUser      *shared_types.User
		inputUser      *shared_types.User
		request        types.ChangePasswordRequest
		storageError   error
		expectedError  error
		updateErrorSet bool
	}{
		{
			name:           "empty reset token",
			setupUser:      nil,
			inputUser:      &shared_types.User{ResetToken: "", Email: "user@example.com"},
			request:        types.ChangePasswordRequest{OldPassword: "old", NewPassword: "new"},
			storageError:   nil,
			expectedError:  types.ErrInvalidResetToken,
			updateErrorSet: false,
		},
		{
			name:           "storage error on GetResetToken",
			setupUser:      nil,
			inputUser:      &shared_types.User{ResetToken: "some-token", Email: "user@example.com"},
			request:        types.ChangePasswordRequest{OldPassword: "old", NewPassword: "new"},
			storageError:   errors.New("database error"),
			expectedError:  types.ErrInvalidResetToken,
			updateErrorSet: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStorage.ExpectedCalls = nil
			mockStorage.On("GetResetToken", test.inputUser.ResetToken).Return(test.setupUser, test.storageError)
			if test.updateErrorSet {
				mockStorage.On("UpdateUser", mock.Anything).Return(types.ErrFailedToUpdateUser)
			} else {
				mockStorage.On("UpdateUser", mock.Anything).Return(nil)
			}
			err := authService.ResetPassword(test.inputUser, test.request)
			assert.Equal(t, test.expectedError, err)
		})
	}
}
