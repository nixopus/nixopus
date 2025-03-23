package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/service"
	"github.com/stretchr/testify/assert"
)

func TestNotificationService_DeleteSmtp(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		mockErr error
		wantErr bool
	}{
		{
			name:    "valid ID and no error from storage layer",
			id:      "123",
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "invalid ID and error from storage layer",
			id:      "invalid",
			mockErr: errors.New("storage error"),
			wantErr: true,
		},
		{
			name:    "empty ID and error from storage layer",
			id:      "",
			mockErr: errors.New("storage error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &MockNotificationStorage{}
			mockStorage.On("DeleteSmtp", tt.id).Return(tt.mockErr)

			s := service.NewNotificationService(nil, context.Background(), logger.NewLogger(), mockStorage)

			err := s.DeleteSmtp(tt.id)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
