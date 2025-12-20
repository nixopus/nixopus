package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// UpdateGithubConnectorRequest updates the GitHub connector request for the given user ID.
//
// If ConnectorID is provided, it updates that specific connector.
// Otherwise, it finds the connector without an installation_id and updates that one.
// If multiple connectors exist and ConnectorID is not provided, it returns an error to prevent ambiguity.
// If no connector without installation_id is found, it updates the first connector (backward compatibility).
//
// If any errors occur during the update process, the method returns the error.
func (c *GithubConnectorService) UpdateGithubConnectorRequest(InstallationID string, UserID string, ConnectorID string) error {
	connectors, err := c.storage.GetAllConnectors(UserID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if len(connectors) == 0 {
		fmt.Println("no connector found")
		return fmt.Errorf("no connectors found for user")
	}

	var connectorToUpdate *shared_types.GithubConnector

	// If ConnectorID is provided, find and update that specific connector
	if ConnectorID != "" {
		// Validate UUID format
		if _, err := uuid.Parse(ConnectorID); err != nil {
			return fmt.Errorf("invalid connector_id format: %v", err)
		}

		// Find the connector with matching ID
		for i := range connectors {
			if connectors[i].ID.String() == ConnectorID {
				connectorToUpdate = &connectors[i]
				break
			}
		}

		if connectorToUpdate == nil {
			return fmt.Errorf("connector with id %s not found", ConnectorID)
		}
	} else {
		// If multiple connectors exist, connector_id is required to avoid ambiguity
		if len(connectors) > 1 {
			return fmt.Errorf("connector_id is required when multiple connectors exist")
		}

		// Find connector without installation_id (newly created connector)
		for i := range connectors {
			if connectors[i].InstallationID == "" || strings.TrimSpace(connectors[i].InstallationID) == "" {
				connectorToUpdate = &connectors[i]
				break
			}
		}

		// If no connector without installation_id found, use first connector (backward compatibility for single connector)
		if connectorToUpdate == nil {
			connectorToUpdate = &connectors[0]
		}
	}

	err = c.storage.UpdateConnector(connectorToUpdate.ID.String(), InstallationID)
	if err != nil {
		return err
	}
	return nil
}
