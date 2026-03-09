package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// CreateConnector creates a new GitHub connector for the given user.
//
// If the request includes credentials (AppID, Pem, etc.), those are used directly.
// Otherwise the connector falls back to the shared GitHub App config from environment variables.
//
// If InstallationID is provided in the request it is set on the connector immediately;
// otherwise it defaults to empty and can be set later via UpdateGithubConnectorRequest.
func (s *GithubConnectorService) CreateConnector(connector *types.CreateGithubConnectorRequest, userID string) error {
	if _, err := uuid.Parse(userID); err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return err
	}

	githubConfig := config.AppConfig.GitHub

	// Use provided credentials if available, otherwise use shared config
	appID := connector.AppID
	slug := connector.Slug
	pem := connector.Pem
	clientID := connector.ClientID
	clientSecret := connector.ClientSecret
	webhookSecret := connector.WebhookSecret

	if appID == "" || pem == "" {
		if githubConfig.AppID == "" || githubConfig.Pem == "" {
			s.logger.Log(logger.Error, "GitHub App credentials not configured", "")
			return fmt.Errorf("GitHub App credentials not configured")
		}
		appID = githubConfig.AppID
		slug = githubConfig.Slug
		pem = githubConfig.Pem
		clientID = githubConfig.ClientID
		clientSecret = githubConfig.ClientSecret
		webhookSecret = githubConfig.WebhookSecret
	}

	installationID := connector.InstallationID

	new_connector := shared_types.GithubConnector{
		ID:             uuid.New(),
		InstallationID: installationID,
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
	err := s.storage.CreateConnector(&new_connector)
	return err
}
