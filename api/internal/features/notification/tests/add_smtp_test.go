package tests

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/net/context"
)

func TestAddSmtp(t *testing.T) {
	tests := []struct {
		name          string
		smtpConfigs   notification.CreateSMTPConfigRequest
		userID        uuid.UUID
		expectedError error
	}{
		{
			name: "valid SMTP configuration and user ID",
			smtpConfigs: notification.CreateSMTPConfigRequest{
				Host:     "example.com",
				Port:     587,
				Username: "user",
				Password: "password",
			},
			userID:        uuid.New(),
			expectedError: nil,
		},
		{
			name: "invalid SMTP configuration (missing host)",
			smtpConfigs: notification.CreateSMTPConfigRequest{
				Port:     587,
				Username: "user",
				Password: "password",
			},
			userID:        uuid.New(),
			expectedError: notification.ErrMissingHost,
		},
		{
			name: "invalid SMTP configuration (missing port)",
			smtpConfigs: notification.CreateSMTPConfigRequest{
				Host:     "example.com",
				Username: "user",
				Password: "password",
			},
			userID:        uuid.New(),
			expectedError: notification.ErrMissingPort,
		},
		{
			name: "invalid SMTP configuration (missing username)",
			smtpConfigs: notification.CreateSMTPConfigRequest{
				Host:     "example.com",
				Port:     587,
				Password: "password",
			},
			userID:        uuid.New(),
			expectedError: notification.ErrMissingUsername,
		},
		{
			name: "invalid SMTP configuration (missing password)",
			smtpConfigs: notification.CreateSMTPConfigRequest{
				Host:     "example.com",
				Port:     587,
				Username: "user",
			},
			userID:        uuid.New(),
			expectedError: notification.ErrMissingPassword,
		},
		{
			name: "invalid user ID (empty UUID)",
			smtpConfigs: notification.CreateSMTPConfigRequest{
				Host:     "example.com",
				Port:     587,
				Username: "user",
				Password: "password",
			},
			userID:        uuid.UUID{},
			expectedError: notification.ErrInvalidRequestType,
		},
		{
			name: "storage operation failure",
			smtpConfigs: notification.CreateSMTPConfigRequest{
				Host:     "example.com",
				Port:     587,
				Username: "user",
				Password: "password",
			},
			userID:        uuid.New(),
			expectedError: errors.New("storage error"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockStorage := &MockNotificationStorage{}
			mockStorage.On("AddSmtp", mock.Anything).Return(test.expectedError)

			service := service.NewNotificationService(nil, context.Background(), logger.NewLogger(), mockStorage)

			err := service.AddSmtp(test.smtpConfigs, test.userID)

			assert.Equal(t, test.expectedError, err)
		})
	}
}