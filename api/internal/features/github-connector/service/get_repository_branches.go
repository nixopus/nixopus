package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

var githubAPIBaseURL = "https://api.github.com"

func SetGithubAPIBaseURL(url string) {
	githubAPIBaseURL = url
}

func (c *GithubConnectorService) GetGithubRepositoryBranches(user_id string, repository_name string) ([]shared_types.GithubRepositoryBranch, error) {
	connectors, err := c.storage.GetAllConnectors(user_id)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	if len(connectors) == 0 {
		c.logger.Log(logger.Error, "No connectors found for user", user_id)
		return []shared_types.GithubRepositoryBranch{}, nil
	}

	installation_id := connectors[0].InstallationID

	jwt := GenerateJwt(&connectors[0])
	if jwt == "" {
		c.logger.Log(logger.Error, "Failed to generate app JWT", "")
		return nil, fmt.Errorf("failed to generate app JWT")
	}

	accessToken, err := c.getInstallationToken(jwt, installation_id)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to get installation token: %s", err.Error()), "")
		return nil, err
	}

	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/repos/%s/branches", githubAPIBaseURL, repository_name), nil)
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

	var branches []shared_types.GithubRepositoryBranch

	err = json.NewDecoder(resp.Body).Decode(&branches)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	return branches, nil
}
