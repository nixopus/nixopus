package service

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetConnector retrieves a GitHub connector by its unique identifier.
//
// This method queries the storage to find a GitHub connector associated with
// the provided ConnectorID. If found, it returns the connector object; otherwise,
// it returns an error.
//
// Parameters:
//
//	ConnectorID - the unique identifier of the GitHub connector to retrieve.
//
// Returns:
//
//	*shared_types.GithubConnector - a pointer to the GitHub connector object if found.
//	error - an error if the connector cannot be retrieved or does not exist.
func (c *GithubConnectorService) GetConnector(ConnectorID string) (*shared_types.GithubConnector, error) {
	connector, err := c.storage.GetConnector(ConnectorID)
	return connector, err
}

// GetAllConnectors retrieves all GitHub connectors associated with the provided UserID.
//
// This method queries the storage to find all GitHub connectors associated with
// the provided UserID. If found, it returns a slice of connector objects;
// otherwise, it returns an error.
//
// Parameters:
//
//	UserID - the unique identifier of the user whose connectors to retrieve.
//
// Returns:
//
//	[]shared_types.GithubConnector - a slice of GitHub connector objects if found.
//	error - an error if the connectors cannot be retrieved or do not exist.
func (c *GithubConnectorService) GetAllConnectors(UserID string) ([]shared_types.GithubConnector, error) {
	connectors, err := c.storage.GetAllConnectors(UserID)
	return connectors, err
}
