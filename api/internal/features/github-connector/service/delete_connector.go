package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// DeleteConnector deletes a GitHub connector for the given user.
//
// This method performs a soft delete on the connector by setting its DeletedAt field.
// It verifies that the connector belongs to the user before deletion.
//
// Parameters:
//
//	ConnectorID - the unique identifier of the connector to delete.
//	UserID - the unique identifier of the user who owns the connector.
//
// Returns:
//
//	error - an error if the connector cannot be deleted or does not exist.
func (c *GithubConnectorService) DeleteConnector(ConnectorID string, UserID string) error {
	// Verify connector exists and belongs to user
	connector, err := c.storage.GetConnector(ConnectorID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return types.ErrConnectorDoesNotExist
	}

	if connector.UserID.String() != UserID {
		c.logger.Log(logger.Error, "User does not own this connector", "")
		return types.ErrPermissionDenied
	}

	// Check if connector is already deleted
	if connector.DeletedAt != nil {
		c.logger.Log(logger.Error, "Connector already deleted", "")
		return types.ErrConnectorDoesNotExist
	}

	// Perform soft delete
	err = c.storage.DeleteConnector(ConnectorID, UserID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return err
	}

	return nil
}
