package gh

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/testutil"
	gctypes "github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ---------------------------------------------------------------------------
// CreateConnector
// ---------------------------------------------------------------------------

func TestCreateConnector_Success(t *testing.T) {
	userID := uuid.New().String()
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("CreateConnector", mock.AnythingOfType("*types.GithubConnector")).Return(nil)

	api := newAPI(mockSt)
	err := api.CreateConnector(&gctypes.CreateGithubConnectorRequest{
		AppID: "app-id", Slug: "slug", Pem: "pem",
		ClientID: "cid", ClientSecret: "cs", WebhookSecret: "ws",
	}, userID)
	assert.NoError(t, err)
	mockSt.AssertExpectations(t)
}

func TestCreateConnector_InvalidUUID(t *testing.T) {
	api := newAPI(testutil.NewMockGithubConnectorStorage())
	err := api.CreateConnector(&gctypes.CreateGithubConnectorRequest{AppID: "a", Pem: "p"}, "not-a-uuid")
	assert.Error(t, err)
}

func TestCreateConnector_StorageError(t *testing.T) {
	userID := uuid.New().String()
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("CreateConnector", mock.AnythingOfType("*types.GithubConnector")).Return(assert.AnError)

	api := newAPI(mockSt)
	err := api.CreateConnector(&gctypes.CreateGithubConnectorRequest{AppID: "a", Pem: "p", Slug: "s"}, userID)
	assert.Error(t, err)
	mockSt.AssertExpectations(t)
}

func TestCreateConnector_NoCredentials_NoConfig(t *testing.T) {
	userID := uuid.New().String()
	api := newAPI(testutil.NewMockGithubConnectorStorage())
	err := api.CreateConnector(&gctypes.CreateGithubConnectorRequest{}, userID)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "GitHub App credentials not configured")
}

func TestCreateConnector_FieldsSet(t *testing.T) {
	userID := uuid.New().String()
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("CreateConnector", mock.AnythingOfType("*types.GithubConnector")).Return(nil)

	req := &gctypes.CreateGithubConnectorRequest{
		AppID: "my-app", Slug: "my-slug", Pem: "my-pem",
		ClientID: "cid", ClientSecret: "csec", WebhookSecret: "whsec",
	}
	api := newAPI(mockSt)
	assert.NoError(t, api.CreateConnector(req, userID))

	created := mockSt.LastConnector()
	assert.Equal(t, req.AppID, created.AppID)
	assert.Equal(t, req.Slug, created.Slug)
	assert.Equal(t, req.Pem, created.Pem)
	assert.Equal(t, req.ClientID, created.ClientID)
	assert.Equal(t, req.ClientSecret, created.ClientSecret)
	assert.Equal(t, req.WebhookSecret, created.WebhookSecret)
	assert.Equal(t, "", created.InstallationID)
	assert.Equal(t, uuid.MustParse(userID), created.UserID)
	assert.WithinDuration(t, time.Now(), created.CreatedAt, 5*time.Second)
	assert.NotEqual(t, uuid.Nil, created.ID)
}

// ---------------------------------------------------------------------------
// UpdateConnectorInstallation
// ---------------------------------------------------------------------------

func TestUpdateConnectorInstallation_NoConnectors(t *testing.T) {
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", "user-123").Return([]shared_types.GithubConnector{}, nil)

	api := newAPI(mockSt)
	err := api.UpdateConnectorInstallation("inst-123", "user-123", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no connectors found for user")
	mockSt.AssertExpectations(t)
}

func TestUpdateConnectorInstallation_GetConnectorsError(t *testing.T) {
	mockSt := testutil.NewMockGithubConnectorStorageWithErr()
	mockSt.On("GetAllConnectors", "user-123").Return(nil, errors.New("failed to get all connectors"))

	api := NewAPI(mockSt, testutil.NewLogger())
	err := api.UpdateConnectorInstallation("inst-123", "user-123", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get all connectors")
	mockSt.AssertExpectations(t)
}

func TestUpdateConnectorInstallation_UpdateError(t *testing.T) {
	connectorID := uuid.New()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	connectors := []shared_types.GithubConnector{{ID: connectorID, UserID: userID, AppID: "app"}}

	mockSt := testutil.NewMockGithubConnectorStorageWithErr()
	mockSt.On("GetAllConnectors", userID.String()).Return(connectors, nil)
	mockSt.On("UpdateConnector", connectorID.String(), "inst-123").Return(errors.New("failed to update connector"))

	api := NewAPI(mockSt, testutil.NewLogger())
	err := api.UpdateConnectorInstallation("inst-123", userID.String(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to")
	mockSt.AssertExpectations(t)
}

func TestUpdateConnectorInstallation_Success(t *testing.T) {
	connectorID := uuid.New()
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	connectors := []shared_types.GithubConnector{{ID: connectorID, UserID: userID, AppID: "app"}}

	customMock := &testutil.CustomMockStorage{}
	customMock.ExpectGetAllConnectors(userID.String(), connectors, nil)
	customMock.ExpectUpdateConnector(connectorID.String(), "inst-123", nil)

	api := NewAPI(customMock, testutil.NewLogger())
	assert.NoError(t, api.UpdateConnectorInstallation("inst-123", userID.String(), ""))
	customMock.VerifyExpectations(t)
}

func TestUpdateConnectorInstallation_WithConnectorID_Valid(t *testing.T) {
	connectorID := uuid.New()
	userID := uuid.New()
	connectors := []shared_types.GithubConnector{{ID: connectorID, UserID: userID}}
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return(connectors, nil)
	mockSt.On("UpdateConnector", connectorID.String(), "inst-456").Return(nil)

	api := newAPI(mockSt)
	assert.NoError(t, api.UpdateConnectorInstallation("inst-456", userID.String(), connectorID.String()))
	mockSt.AssertExpectations(t)
}

func TestUpdateConnectorInstallation_WithConnectorID_InvalidUUID(t *testing.T) {
	userID := uuid.New()
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{{ID: uuid.New(), UserID: userID}}, nil)

	api := newAPI(mockSt)
	err := api.UpdateConnectorInstallation("inst", userID.String(), "not-a-uuid")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid connector_id format")
	mockSt.AssertExpectations(t)
}

func TestUpdateConnectorInstallation_WithConnectorID_NotFound(t *testing.T) {
	userID := uuid.New()
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return([]shared_types.GithubConnector{{ID: uuid.New(), UserID: userID}}, nil)

	api := newAPI(mockSt)
	err := api.UpdateConnectorInstallation("inst", userID.String(), uuid.New().String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	mockSt.AssertExpectations(t)
}

func TestUpdateConnectorInstallation_MultipleConnectors_RequiresID(t *testing.T) {
	userID := uuid.New()
	connectors := []shared_types.GithubConnector{{ID: uuid.New(), UserID: userID}, {ID: uuid.New(), UserID: userID}}
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return(connectors, nil)

	api := newAPI(mockSt)
	err := api.UpdateConnectorInstallation("inst", userID.String(), "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connector_id is required when multiple connectors exist")
	mockSt.AssertExpectations(t)
}

func TestUpdateConnectorInstallation_SingleConnector_UsesFirst(t *testing.T) {
	connectorID := uuid.New()
	userID := uuid.New()
	connectors := []shared_types.GithubConnector{{ID: connectorID, UserID: userID, InstallationID: "existing"}}
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID.String()).Return(connectors, nil)
	mockSt.On("UpdateConnector", connectorID.String(), "new-inst").Return(nil)

	api := newAPI(mockSt)
	assert.NoError(t, api.UpdateConnectorInstallation("new-inst", userID.String(), ""))
	mockSt.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// DeleteConnector
// ---------------------------------------------------------------------------

func TestDeleteConnector_ConnectorNotFound(t *testing.T) {
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetConnector", "c1").Return((*shared_types.GithubConnector)(nil), errors.New("not found"))

	api := newAPI(mockSt)
	err := api.DeleteConnector("c1", "u1")
	assert.ErrorIs(t, err, gctypes.ErrConnectorDoesNotExist)
	mockSt.AssertExpectations(t)
}

func TestDeleteConnector_WrongUser(t *testing.T) {
	ownerID := uuid.New()
	connector := &shared_types.GithubConnector{ID: uuid.New(), UserID: ownerID}
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetConnector", connector.ID.String()).Return(connector, nil)

	api := newAPI(mockSt)
	err := api.DeleteConnector(connector.ID.String(), uuid.New().String())
	assert.ErrorIs(t, err, gctypes.ErrPermissionDenied)
	mockSt.AssertExpectations(t)
}

func TestDeleteConnector_AlreadyDeleted(t *testing.T) {
	userID := uuid.New()
	now := time.Now()
	connector := &shared_types.GithubConnector{ID: uuid.New(), UserID: userID, DeletedAt: &now}
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetConnector", connector.ID.String()).Return(connector, nil)

	api := newAPI(mockSt)
	err := api.DeleteConnector(connector.ID.String(), userID.String())
	assert.ErrorIs(t, err, gctypes.ErrConnectorDoesNotExist)
	mockSt.AssertExpectations(t)
}

func TestDeleteConnector_StorageError(t *testing.T) {
	userID := uuid.New()
	connector := &shared_types.GithubConnector{ID: uuid.New(), UserID: userID}
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetConnector", connector.ID.String()).Return(connector, nil)
	mockSt.On("DeleteConnector", connector.ID.String(), userID.String()).Return(errors.New("db error"))

	api := newAPI(mockSt)
	assert.EqualError(t, api.DeleteConnector(connector.ID.String(), userID.String()), "db error")
	mockSt.AssertExpectations(t)
}

func TestDeleteConnector_Success(t *testing.T) {
	userID := uuid.New()
	connector := &shared_types.GithubConnector{ID: uuid.New(), UserID: userID}
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetConnector", connector.ID.String()).Return(connector, nil)
	mockSt.On("DeleteConnector", connector.ID.String(), userID.String()).Return(nil)

	api := newAPI(mockSt)
	assert.NoError(t, api.DeleteConnector(connector.ID.String(), userID.String()))
	mockSt.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// GetConnector / GetAllConnectors
// ---------------------------------------------------------------------------

func TestGetConnector_Passthrough(t *testing.T) {
	connector := &shared_types.GithubConnector{ID: uuid.New()}
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetConnector", connector.ID.String()).Return(connector, nil)

	api := newAPI(mockSt)
	got, err := api.GetConnector(connector.ID.String())
	assert.NoError(t, err)
	assert.Equal(t, connector, got)
	mockSt.AssertExpectations(t)
}

func TestGetConnector_Error(t *testing.T) {
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetConnector", "bad").Return((*shared_types.GithubConnector)(nil), errors.New("not found"))

	api := newAPI(mockSt)
	_, err := api.GetConnector("bad")
	assert.Error(t, err)
	mockSt.AssertExpectations(t)
}

func TestGetAllConnectors_Passthrough(t *testing.T) {
	userID := uuid.New().String()
	connectors := []shared_types.GithubConnector{{ID: uuid.New()}}
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", userID).Return(connectors, nil)

	api := newAPI(mockSt)
	got, err := api.GetAllConnectors(userID)
	assert.NoError(t, err)
	assert.Equal(t, connectors, got)
	mockSt.AssertExpectations(t)
}

func TestGetAllConnectors_Error(t *testing.T) {
	mockSt := testutil.NewMockGithubConnectorStorage()
	mockSt.On("GetAllConnectors", "u1").Return(nil, errors.New("db error"))

	api := newAPI(mockSt)
	_, err := api.GetAllConnectors("u1")
	assert.Error(t, err)
	mockSt.AssertExpectations(t)
}

// ---------------------------------------------------------------------------
// CreateAuthenticatedRepoURL
// ---------------------------------------------------------------------------

func TestCreateAuthenticatedRepoURL(t *testing.T) {
	tests := []struct {
		name        string
		repoURL     string
		accessToken string
		expectedURL string
		wantErr     string
	}{
		{name: "Valid HTTPS URL", repoURL: "https://github.com/user/repo", accessToken: "token", expectedURL: "https://oauth2:token@github.com/user/repo"},
		{name: "Valid SSH URL", repoURL: "git@github.com:user/repo.git", accessToken: "token", expectedURL: "https://oauth2:token@github.com/user/repo.git"},
		{name: "Invalid URL format", repoURL: "invalid-url", wantErr: "unsupported repository URL format"},
		{name: "Unsupported scheme", repoURL: "ftp://github.com/user/repo", wantErr: "unsupported repository URL format"},
		{name: "Empty URL", repoURL: "", wantErr: "unsupported repository URL format"},
		{name: "Empty token", repoURL: "https://github.com/user/repo", expectedURL: "https://oauth2:@github.com/user/repo"},
		{name: "SSH URL missing slash", repoURL: "git@github.com:noslash", wantErr: "invalid SSH repository URL format"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CreateAuthenticatedRepoURL(tt.repoURL, tt.accessToken)
			if tt.wantErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				assert.Empty(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedURL, got)
			}
		})
	}
}
