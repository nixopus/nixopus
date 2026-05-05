package gh

import (
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	gctypes "github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// CreateConnector creates a new GitHub connector for the given user.
func (a *API) CreateConnector(connector *gctypes.CreateGithubConnectorRequest, userID string) error {
	if _, err := uuid.Parse(userID); err != nil {
		a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: CreateConnector invalid user_id: %v", err), userID)
		return err
	}
	ctxStr := fmt.Sprintf("user_id=%s", userID)
	githubConfig := config.AppConfig.GitHub
	appID := connector.AppID
	slug := connector.Slug
	pem := connector.Pem
	clientID := connector.ClientID
	clientSecret := connector.ClientSecret
	webhookSecret := connector.WebhookSecret
	if appID == "" || pem == "" {
		if githubConfig.AppID == "" || githubConfig.Pem == "" {
			a.Logger.Log(logger.Error, "github connector service: CreateConnector GitHub App credentials not configured", ctxStr)
			return fmt.Errorf("GitHub App credentials not configured")
		}
		appID = githubConfig.AppID
		slug = githubConfig.Slug
		pem = githubConfig.Pem
		clientID = githubConfig.ClientID
		clientSecret = githubConfig.ClientSecret
		webhookSecret = githubConfig.WebhookSecret
	}
	newConn := shared_types.GithubConnector{
		ID:             uuid.New(),
		InstallationID: connector.InstallationID,
		AppID:          appID,
		Slug:           slug,
		Pem:            pem,
		ClientID:       clientID,
		ClientSecret:   clientSecret,
		WebhookSecret:  webhookSecret,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeletedAt:      nil,
		UserID:         uuid.MustParse(userID),
	}
	if err := a.Storage.CreateConnector(&newConn); err != nil {
		a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: CreateConnector: %v", err), ctxStr)
		return err
	}
	a.Logger.Log(logger.Info, "github connector service: CreateConnector ok", fmt.Sprintf("%s connector_id=%s", ctxStr, newConn.ID))
	return nil
}

// UpdateConnectorInstallation updates the GitHub App installation ID for the user's connector.
func (a *API) UpdateConnectorInstallation(installationID, userID, connectorID string) error {
	ctxStr := fmt.Sprintf("user_id=%s connector_id=%s", userID, connectorID)
	connectors, err := a.Storage.GetAllConnectors(userID)
	if err != nil {
		a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: UpdateConnectorInstallation list: %v", err), ctxStr)
		return err
	}
	if len(connectors) == 0 {
		a.Logger.Log(logger.Error, "github connector service: UpdateConnectorInstallation no connectors for user", ctxStr)
		return fmt.Errorf("no connectors found for user")
	}
	var connectorToUpdate *shared_types.GithubConnector
	if connectorID != "" {
		if _, err := uuid.Parse(connectorID); err != nil {
			a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: UpdateConnectorInstallation invalid connector_id: %v", err), ctxStr)
			return fmt.Errorf("invalid connector_id format: %v", err)
		}
		for i := range connectors {
			if connectors[i].ID.String() == connectorID {
				connectorToUpdate = &connectors[i]
				break
			}
		}
		if connectorToUpdate == nil {
			a.Logger.Log(logger.Error, "github connector service: UpdateConnectorInstallation connector not found for user", ctxStr)
			return fmt.Errorf("connector with id %s not found", connectorID)
		}
	} else {
		if len(connectors) > 1 {
			a.Logger.Log(logger.Error, "github connector service: UpdateConnectorInstallation connector_id required when multiple connectors", ctxStr)
			return fmt.Errorf("connector_id is required when multiple connectors exist")
		}
		for i := range connectors {
			if connectors[i].InstallationID == "" || strings.TrimSpace(connectors[i].InstallationID) == "" {
				connectorToUpdate = &connectors[i]
				break
			}
		}
		if connectorToUpdate == nil {
			connectorToUpdate = &connectors[0]
		}
	}
	ctxStr = fmt.Sprintf("%s target_connector_id=%s", ctxStr, connectorToUpdate.ID)
	if err := a.Storage.UpdateConnector(connectorToUpdate.ID.String(), installationID); err != nil {
		a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: UpdateConnectorInstallation: %v", err), ctxStr)
		return err
	}
	a.Logger.Log(logger.Info, "github connector service: UpdateConnectorInstallation ok", ctxStr)
	return nil
}

// DeleteConnector soft-deletes a connector after ownership checks.
func (a *API) DeleteConnector(connectorID, userID string) error {
	ctxStr := fmt.Sprintf("connector_id=%s user_id=%s", connectorID, userID)
	connector, err := a.Storage.GetConnector(connectorID)
	if err != nil {
		a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: DeleteConnector lookup: %v", err), ctxStr)
		return gctypes.ErrConnectorDoesNotExist
	}
	if connector.UserID.String() != userID {
		a.Logger.Log(logger.Error, "github connector service: DeleteConnector permission denied", ctxStr)
		return gctypes.ErrPermissionDenied
	}
	if connector.DeletedAt != nil {
		a.Logger.Log(logger.Error, "github connector service: DeleteConnector already deleted", ctxStr)
		return gctypes.ErrConnectorDoesNotExist
	}
	err = a.Storage.DeleteConnector(connectorID, userID)
	if err != nil {
		a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: DeleteConnector: %v", err), ctxStr)
		return err
	}
	a.Logger.Log(logger.Info, "github connector service: DeleteConnector ok", ctxStr)
	return nil
}

// GetConnector returns a connector by ID.
func (a *API) GetConnector(connectorID string) (*shared_types.GithubConnector, error) {
	return a.Storage.GetConnector(connectorID)
}

// GetAllConnectors lists connectors for a user.
func (a *API) GetAllConnectors(userID string) ([]shared_types.GithubConnector, error) {
	return a.Storage.GetAllConnectors(userID)
}

// CreateAuthenticatedRepoURL builds an HTTPS URL with embedded token for git operations.
func CreateAuthenticatedRepoURL(repoURL, accessToken string) (string, error) {
	if strings.HasPrefix(repoURL, "https://") {
		parsedURL, err := url.Parse(repoURL)
		if err != nil {
			return "", fmt.Errorf("invalid repository URL: %w", err)
		}
		return fmt.Sprintf("https://oauth2:%s@github.com%s", accessToken, parsedURL.Path), nil
	}
	if strings.HasPrefix(repoURL, "git@github.com") {
		parts := strings.Split(strings.TrimPrefix(repoURL, "git@github.com:"), "/")
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid SSH repository URL format")
		}
		owner := parts[0]
		repo := strings.TrimSuffix(parts[len(parts)-1], ".git")
		return fmt.Sprintf("https://oauth2:%s@github.com/%s/%s.git", accessToken, owner, repo), nil
	}
	return "", fmt.Errorf("unsupported repository URL format")
}
