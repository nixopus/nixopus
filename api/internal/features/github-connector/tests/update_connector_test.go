package tests

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func TestUpdateGithubConnectorRequest_NoConnectorsFound(t *testing.T) {
	storage := NewMockGithubConnectorStorage()
	storage.On("GetAllConnectors", "user-123").Return([]shared_types.GithubConnector{}, nil)

	svc := service.NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), storage)
	err := svc.UpdateGithubConnectorRequest("installation-123", "user-123")

	assert.Nil(t, err)
	storage.AssertExpectations(t)
}

func TestUpdateGithubConnectorRequest_UpdateConnectorError(t *testing.T) {
	storage := NewMockGithubConnectorStorageWithErr()

	connectorID := uuid.New()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	connectors := []shared_types.GithubConnector{
		{
			ID:     connectorID,
			UserID: userID,
			AppID:  "test-app-id",
		},
	}

	storage.On("GetAllConnectors", userID.String()).Return(connectors, nil)
	storage.On("UpdateConnector", connectorID.String(), "installation-123").Return(errors.New("failed to update connector"))

	svc := service.NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), storage)
	err := svc.UpdateGithubConnectorRequest("installation-123", userID.String())

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to")
	storage.AssertExpectations(t)
}

func TestUpdateGithubConnectorRequest_GetAllConnectorsError(t *testing.T) {
	storage := NewMockGithubConnectorStorageWithErr()
	storage.On("GetAllConnectors", "user-123").Return(nil, errors.New("failed to get all connectors"))

	svc := service.NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), storage)
	err := svc.UpdateGithubConnectorRequest("installation-123", "user-123")

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "failed to get all connectors")
	storage.AssertExpectations(t)
}

func TestUpdateGithubConnectorRequest_SuccessfulUpdate(t *testing.T) {
	customMock := &CustomMockStorage{}

	connectorID := uuid.New()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	connector := shared_types.GithubConnector{
		ID:     connectorID,
		UserID: userID,
		AppID:  "test-app-id",
	}
	connectors := []shared_types.GithubConnector{connector}

	customMock.ExpectGetAllConnectors(userID.String(), connectors, nil)
	customMock.ExpectUpdateConnector(connectorID.String(), "installation-123", nil)

	svc := service.NewGithubConnectorService(nil, context.Background(), logger.NewLogger(), customMock)
	err := svc.UpdateGithubConnectorRequest("installation-123", userID.String())

	assert.Nil(t, err)
	customMock.VerifyExpectations(t)
}
