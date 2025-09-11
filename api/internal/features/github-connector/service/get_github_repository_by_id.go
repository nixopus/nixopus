package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetGithubRepositoryByID gets repository details from GitHub by repository ID.
func (c *GithubConnectorService) GetGithubRepositoryByID(userID string, repoID uint64) (*shared_types.GithubRepository, error) {
	connectors, err := c.storage.GetAllConnectors(userID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	if len(connectors) == 0 {
		c.logger.Log(logger.Error, "No connectors found for user", userID)
		return nil, fmt.Errorf("no connectors found for user")
	}

	installationID := connectors[0].InstallationID
	jwt := GenerateJwt(&connectors[0])
	if jwt == "" {
		c.logger.Log(logger.Error, "Failed to generate app JWT", "")
		return nil, fmt.Errorf("failed to generate app JWT")
	}

	accessToken, err := c.getInstallationToken(jwt, installationID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to get installation token: %s", err.Error()), "")
		return nil, err
	}

	client := &http.Client{}
	url := fmt.Sprintf("%s/repositories/%d", githubAPIBaseURL, repoID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nixopus")

	resp, err := client.Do(req)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.logger.Log(logger.Error, fmt.Sprintf("GitHub API error: %s - %s", resp.Status, string(bodyBytes)), "")
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var repo shared_types.GithubRepository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	return &repo, nil
}
