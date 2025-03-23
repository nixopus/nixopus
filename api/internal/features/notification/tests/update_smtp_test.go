package tests

import (
	"errors"
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateSmtp(t *testing.T) {
	tests := []struct {
		name        string
		setupMock   func() *MockNotificationStorage
		SMTPConfigs notification.UpdateSMTPConfigRequest
		wantErr     bool
	}{
		{
			name: "successful update",
			setupMock: func() *MockNotificationStorage {
				mockStorage := new(MockNotificationStorage)
				mockStorage.On("UpdateSmtp", mock.AnythingOfType("*notification.UpdateSMTPConfigRequest")).Return(nil)
				return mockStorage
			},
			SMTPConfigs: notification.UpdateSMTPConfigRequest{
				Host: "smtp.example.com",
				Port: 587,
			},
			wantErr: false,
		},
		{
			name: "error from storage layer",
			setupMock: func() *MockNotificationStorage {
				mockStorage := new(MockNotificationStorage)
				mockStorage.On("UpdateSmtp", mock.AnythingOfType("*notification.UpdateSMTPConfigRequest")).Return(errors.New("storage error"))
				return mockStorage
			},
			SMTPConfigs: notification.UpdateSMTPConfigRequest{
				Host: "smtp.example.com",
				Port: 587,
			},
			wantErr: true,
		},
		{
			name: "invalid SMTPConfigs - empty host",
			setupMock: func() *MockNotificationStorage {
				return new(MockNotificationStorage)
			},
			SMTPConfigs: notification.UpdateSMTPConfigRequest{
				Host: "",
				Port: 587,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockStorage *MockNotificationStorage
			if tt.setupMock != nil {
				mockStorage = tt.setupMock()
			}

			s := service.NewNotificationService(nil, nil, logger.NewLogger(), mockStorage)
			err := s.UpdateSmtp(tt.SMTPConfigs)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if mockStorage != nil {
					mockStorage.AssertExpectations(t)
				}
			}
		})
	}
}
