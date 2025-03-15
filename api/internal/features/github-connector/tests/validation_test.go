package tests

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/validation"
	"github.com/stretchr/testify/assert"
)

func TestNewValidator(t *testing.T) {
	mockRepo := NewMockGithubConnectorStorage()
	validator := validation.NewValidator(mockRepo)

	assert.NotNil(t, validator)
}

func TestValidateRequest_InvalidType(t *testing.T) {
	mockRepo := NewMockGithubConnectorStorage()
	validator := validation.NewValidator(mockRepo)
	user := &shared_types.User{}

	err := validator.ValidateRequest("invalid type", user)

	assert.Equal(t, types.ErrInvalidRequestType, err)
}

func TestValidateCreateGithubConnectorRequest(t *testing.T) {
	mockRepo := NewMockGithubConnectorStorage()
	validator := validation.NewValidator(mockRepo)
	user := &shared_types.User{}

	tests := []struct {
		name    string
		request types.CreateGithubConnectorRequest
		wantErr error
	}{
		{
			name: "Valid request",
			request: types.CreateGithubConnectorRequest{
				Slug:          "test-slug",
				Pem:           "test-pem",
				ClientID:      "test-client-id",
				ClientSecret:  "test-client-secret",
				WebhookSecret: "test-webhook-secret",
			},
			wantErr: nil,
		},
		{
			name: "Missing slug",
			request: types.CreateGithubConnectorRequest{
				Pem:           "test-pem",
				ClientID:      "test-client-id",
				ClientSecret:  "test-client-secret",
				WebhookSecret: "test-webhook-secret",
			},
			wantErr: types.ErrMissingSlug,
		},
		{
			name: "Missing pem",
			request: types.CreateGithubConnectorRequest{
				Slug:          "test-slug",
				ClientID:      "test-client-id",
				ClientSecret:  "test-client-secret",
				WebhookSecret: "test-webhook-secret",
			},
			wantErr: types.ErrMissingPem,
		},
		{
			name: "Missing client ID",
			request: types.CreateGithubConnectorRequest{
				Slug:          "test-slug",
				Pem:           "test-pem",
				ClientSecret:  "test-client-secret",
				WebhookSecret: "test-webhook-secret",
			},
			wantErr: types.ErrMissingClientID,
		},
		{
			name: "Missing client secret",
			request: types.CreateGithubConnectorRequest{
				Slug:          "test-slug",
				Pem:           "test-pem",
				ClientID:      "test-client-id",
				WebhookSecret: "test-webhook-secret",
			},
			wantErr: types.ErrMissingClientSecret,
		},
		{
			name: "Missing webhook secret",
			request: types.CreateGithubConnectorRequest{
				Slug:         "test-slug",
				Pem:          "test-pem",
				ClientID:     "test-client-id",
				ClientSecret: "test-client-secret",
			},
			wantErr: types.ErrMissingWebhookSecret,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &tt.request
			err := validator.ValidateRequest(req, user)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidateUpdateGithubConnectorRequest(t *testing.T) {
	userID := uuid.New()
	connectorUserID := userID
	
	adminUser := &shared_types.User{
		ID:   userID,
		Type: "admin",
	}
	
	regularUser := &shared_types.User{
		ID:   userID,
		Type: "regular",
	}
	
	differentUserID := uuid.New()
	differentUser := &shared_types.User{
		ID:   differentUserID,
		Type: "regular",
	}

	connector := shared_types.GithubConnector{
		ID:     uuid.New(),
		UserID: connectorUserID,
		AppID:  "test-app-id",
	}

	tests := []struct {
		name      string
		request   types.UpdateGithubConnectorRequest
		user      *shared_types.User
		mockSetup func(*MockGithubConnectorStorage)
		wantErr   error
	}{
		{
			name: "Valid request same user",
			request: types.UpdateGithubConnectorRequest{
				InstallationID: "test-installation-id",
			},
			user: regularUser,
			mockSetup: func(mockRepo *MockGithubConnectorStorage) {
				mockRepo.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "Valid request admin user",
			request: types.UpdateGithubConnectorRequest{
				InstallationID: "test-installation-id",
			},
			user: adminUser,
			mockSetup: func(mockRepo *MockGithubConnectorStorage) {
				mockRepo.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{connector}, nil).Once()
			},
			wantErr: nil,
		},
		{
			name: "Missing installation ID",
			request: types.UpdateGithubConnectorRequest{},
			user: regularUser,
			mockSetup: func(mockRepo *MockGithubConnectorStorage) {},
			wantErr: types.ErrMissingInstallationID,
		},
		{
			name: "No connectors",
			request: types.UpdateGithubConnectorRequest{
				InstallationID: "test-installation-id",
			},
			user: regularUser,
			mockSetup: func(mockRepo *MockGithubConnectorStorage) {
				mockRepo.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{}, nil).Once()
			},
			wantErr: types.ErrNoConnectors,
		},
		{
			name: "Storage error",
			request: types.UpdateGithubConnectorRequest{
				InstallationID: "test-installation-id",
			},
			user: regularUser,
			mockSetup: func(mockRepo *MockGithubConnectorStorage) {
				mockRepo.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{}, errors.New("storage error")).Once()
			},
			wantErr: errors.New("storage error"),
		},
		{
			name: "Permission denied",
			request: types.UpdateGithubConnectorRequest{
				InstallationID: "test-installation-id",
			},
			user: differentUser,
			mockSetup: func(mockRepo *MockGithubConnectorStorage) {
				mockRepo.On("GetAllConnectors", differentUserID.String()).Return([]shared_types.GithubConnector{connector}, nil).Once()
			},
			wantErr: types.ErrPermissionDenied,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockGithubConnectorStorage()
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}
			
			validator := validation.NewValidator(mockRepo)
			req := &tt.request
			err := validator.ValidateRequest(req, tt.user)
			
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else if errors.Is(err, tt.wantErr) {
				assert.Equal(t, tt.wantErr, err)
			} else if tt.wantErr.Error() != "" && err != nil {
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			}
			
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestValidateUpdateGithubConnectorRequest_WithErrorMock(t *testing.T) {
	userID := uuid.New()
	regularUser := &shared_types.User{
		ID:   userID,
		Type: "regular",
	}
	
	t.Run("Storage error", func(t *testing.T) {
		mockRepo := NewMockGithubConnectorStorageWithErr()
		mockRepo.On("GetAllConnectors", userID.String()).Return(nil, errors.New("failed to get all connectors")).Once()
		
		validator := validation.NewValidator(mockRepo)
		req := &types.UpdateGithubConnectorRequest{
			InstallationID: "test-installation-id",
		}
		
		err := validator.ValidateRequest(req, regularUser)
		assert.Error(t, err)
		assert.Equal(t, "failed to get all connectors", err.Error())
		
		mockRepo.AssertExpectations(t)
	})
}