package tests

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateConnector(t *testing.T) {
	userID := uuid.New().String()
	request := &types.CreateGithubConnectorRequest{
		AppID:          "test-app-id",
		Slug:           "test-slug",
		Pem:            "test-pem",
		ClientID:       "test-client-id",
		ClientSecret:   "test-client-secret",
		WebhookSecret:  "test-webhook-secret",
	}

	tests := []struct {
		name          string
		request       *types.CreateGithubConnectorRequest
		userID        string
		mockSetup     func(*MockGithubConnectorStorage)
		expectedError bool
		expectedErrMsg string
	}{
		{
			name:          "Success case",
			request:       request,
			userID:        userID,
			mockSetup: func(mockRepo *MockGithubConnectorStorage) {
				mockRepo.On("CreateConnector", mock.AnythingOfType("*types.GithubConnector")).Return(nil).Once()
			},
			expectedError: false,
		},
		{
			name:          "Invalid UUID",
			request:       request,
			userID:        "invalid-uuid",
			mockSetup:     func(mockRepo *MockGithubConnectorStorage) {},
			expectedError: true,
			expectedErrMsg: "invalid UUID",
		},
		{
			name:          "Storage error",
			request:       request,
			userID:        userID,
			mockSetup: func(mockRepo *MockGithubConnectorStorage) {
				mockRepo.On("CreateConnector", mock.AnythingOfType("*types.GithubConnector")).Return(assert.AnError).Once()
			},
			expectedError: true,
			expectedErrMsg: "assert.AnError general error for testing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockGithubConnectorStorage()
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			svc := service.NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), mockRepo)

			err := svc.CreateConnector(tt.request, tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
				if tt.expectedErrMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrMsg)
				}
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)

			if !tt.expectedError && mockRepo.methodCalls["CreateConnector"] > 0 {
				createdConnector := mockRepo.lastConnector

				assert.Equal(t, tt.request.AppID, createdConnector.AppID)
				assert.Equal(t, tt.request.Slug, createdConnector.Slug)
				assert.Equal(t, tt.request.Pem, createdConnector.Pem)
				assert.Equal(t, tt.request.ClientID, createdConnector.ClientID)
				assert.Equal(t, tt.request.ClientSecret, createdConnector.ClientSecret)
				assert.Equal(t, tt.request.WebhookSecret, createdConnector.WebhookSecret)
				assert.Equal(t, "", createdConnector.InstallationID)
				assert.Equal(t, uuid.MustParse(tt.userID), createdConnector.UserID)
				
				assert.WithinDuration(t, time.Now(), createdConnector.CreatedAt, 5*time.Second)
				assert.WithinDuration(t, time.Now(), createdConnector.UpdatedAt, 5*time.Second)
				assert.Nil(t, createdConnector.DeletedAt)
				
				assert.NotEqual(t, uuid.Nil, createdConnector.ID)
			}
		})
	}
}