package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *GithubConnectorService) CreateConnector(connector *types.CreateGithubConnectorRequest, userID string) error {
	new_connector := shared_types.GithubConnector{
		ID:             uuid.New(),
		InstallationID: "",
		AppID:          connector.AppID,
		Slug:           connector.Slug,
		Pem:            connector.Pem,
		ClientID:       connector.ClientID,
		ClientSecret:   connector.ClientSecret,
		WebhookSecret:  connector.WebhookSecret,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeletedAt:      nil,
		UserID:         uuid.MustParse(userID),
	}
	err := s.storage.CreateConnector(&new_connector)
	return err
}
