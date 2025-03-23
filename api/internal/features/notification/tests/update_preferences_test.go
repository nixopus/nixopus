package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/service"
	"github.com/stretchr/testify/mock"
)

func TestUpdatePreference(t *testing.T) {
	tests := []struct {
		name        string
		preferences notification.UpdatePreferenceRequest
		userID      uuid.UUID
		storageErr  error
		wantErr     bool
	}{
		{
			name: "valid input",
			preferences: notification.UpdatePreferenceRequest{
				Category: "test-category",
				Type:     "test-type",
				Enabled:  true,
			},
			userID:     uuid.New(),
			storageErr: nil,
			wantErr:    false,
		},
		{
			name: "invalid input - empty category",
			preferences: notification.UpdatePreferenceRequest{
				Category: "",
				Type:     "test-type",
				Enabled:  true,
			},
			userID:     uuid.New(),
			storageErr: nil,
			wantErr:    true,
		},
		{
			name: "invalid input - empty type",
			preferences: notification.UpdatePreferenceRequest{
				Category: "test-category",
				Type:     "",
				Enabled:  true,
			},
			userID:     uuid.New(),
			storageErr: nil,
			wantErr:    true,
		},
		{
			name: "invalid input - empty userID",
			preferences: notification.UpdatePreferenceRequest{
				Category: "test-category",
				Type:     "test-type",
				Enabled:  true,
			},
			userID:     uuid.Nil,
			storageErr: nil,
			wantErr:    true,
		},
		{
			name: "storage error",
			preferences: notification.UpdatePreferenceRequest{
				Category: "test-category",
				Type:     "test-type",
				Enabled:  true,
			},
			userID:     uuid.New(),
			storageErr: errors.New("storage error"),
			wantErr:    true,
		},
	}

	mockStorage := &MockNotificationStorage{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage.On("UpdatePreference", mock.Anything, tt.preferences, tt.userID).Return(tt.storageErr)

			s := service.NewNotificationService(nil, context.Background(), logger.NewLogger(), mockStorage)
			err := s.UpdatePreference(tt.preferences, tt.userID)
			if (err != nil) != tt.wantErr {
				t.Errorf("UpdatePreference() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
