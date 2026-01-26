package service

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// CreateConnector creates a new GitHub connector for the given user.
//
// The connector is created with the InstallationID set to an empty string,
// indicating that it is not associated with any installation yet.
//
// The connector is also created with the CreatedAt, UpdatedAt, and DeletedAt
// fields set to the current time, indicating that the connector is newly
// created.
//
// The UserID field is set to the given userID.
//
// If the connector cannot be created, an error is returned.
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

	new_connector := shared_types.GithubConnector{
		ID:             uuid.New(),
		InstallationID: "",
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
