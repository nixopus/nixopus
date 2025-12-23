package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetGithubRepositoriesPaginated fetches repositories for the user's GitHub installation with pagination.
// If connectorID is provided, it uses that specific connector. Otherwise, it finds a connector with a valid installation_id.
func (c *GithubConnectorService) GetGithubRepositoriesPaginated(userID string, page int, pageSize int, connectorID string) ([]shared_types.GithubRepository, int, error) {
	connectors, err := c.storage.GetAllConnectors(userID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, 0, err
	}

	if len(connectors) == 0 {
		c.logger.Log(logger.Error, "No connectors found for user", userID)
		return []shared_types.GithubRepository{}, 0, nil
	}

	var connectorToUse *shared_types.GithubConnector

	// If connectorID is provided, find that specific connector
	if connectorID != "" {
		for i := range connectors {
			if connectors[i].ID.String() == connectorID {
				connectorToUse = &connectors[i]
				break
			}
		}
		if connectorToUse == nil {
			c.logger.Log(logger.Error, fmt.Sprintf("Connector with id %s not found for user", connectorID), userID)
			return nil, 0, fmt.Errorf("connector not found")
		}
	} else {
		// Find connector with valid installation_id (not empty)
		for i := range connectors {
			if connectors[i].InstallationID != "" && connectors[i].InstallationID != " " {
				connectorToUse = &connectors[i]
				break
			}
		}
		// If no connector with installation_id found, return error
		if connectorToUse == nil {
			c.logger.Log(logger.Error, "No connector with valid installation_id found for user", userID)
			return nil, 0, fmt.Errorf("no connector with valid installation found")
		}
	}

	// Validate installation_id is not empty
	if connectorToUse.InstallationID == "" || connectorToUse.InstallationID == " " {
		c.logger.Log(logger.Error, fmt.Sprintf("Connector %s has empty installation_id", connectorToUse.ID.String()), userID)
		return nil, 0, fmt.Errorf("connector has no installation_id")
	}

	installation_id := connectorToUse.InstallationID
	jwt := GenerateJwt(connectorToUse)

	accessToken, err := c.getInstallationToken(jwt, installation_id)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to get installation token: %s", err.Error()), "")
		return nil, 0, err
	}

	client := &http.Client{}
	url := fmt.Sprintf("https://api.github.com/installation/repositories?per_page=%d&page=%d", pageSize, page)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, 0, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nixopus")

	resp, err := client.Do(req)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.logger.Log(logger.Error, fmt.Sprintf("GitHub API error: %s - %s", resp.Status, string(bodyBytes)), "")
		return nil, 0, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var response struct {
		TotalCount   int                             `json:"total_count"`
		Repositories []shared_types.GithubRepository `json:"repositories"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, 0, err
	}

	return response.Repositories, response.TotalCount, nil
}
