package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
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
