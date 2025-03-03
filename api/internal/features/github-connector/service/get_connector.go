package service

import (
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *GithubConnectorService) GetConnector(ConnectorID string) (*shared_types.GithubConnector, error) {
	connector, err := c.storage.GetConnector(ConnectorID)
	return connector, err
}

func (c *GithubConnectorService) GetAllConnectors(UserID string) ([]shared_types.GithubConnector, error) {
	connectors, err := c.storage.GetAllConnectors(UserID)
	return connectors, err
}