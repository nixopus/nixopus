package service

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/nixopus/nixopus/api/internal/features/logger"
)

type githubContentResponse struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

// GetRepositoryFileContent fetches a single file from a GitHub repository
// using the Contents API. The repository can be a numeric ID or "owner/repo".
func (c *GithubConnectorService) GetRepositoryFileContent(userID string, repository string, branch string, filePath string) ([]byte, error) {
	connectors, err := c.storage.GetAllConnectors(userID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	if len(connectors) == 0 {
		return nil, fmt.Errorf("no GitHub connectors found for user")
	}

	jwt := GenerateJwt(&connectors[0])
	if jwt == "" {
		return nil, fmt.Errorf("failed to generate GitHub App JWT")
	}

	accessToken, err := c.getInstallationToken(jwt, connectors[0].InstallationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get installation token: %w", err)
	}

	var repoFullName string
	if repoID, parseErr := strconv.ParseUint(repository, 10, 64); parseErr == nil {
		repo, err := c.GetGithubRepositoryByID(userID, repoID)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve repository ID %d: %w", repoID, err)
		}
		repoFullName = repo.FullName
	} else {
		repoFullName = repository
	}

	cleanPath := strings.TrimPrefix(filePath, "/")
	apiURL := fmt.Sprintf("%s/repos/%s/contents/%s?ref=%s",
		githubAPIBaseURL, repoFullName, cleanPath, url.QueryEscape(branch))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nixopus")

	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("GitHub API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.logger.Log(logger.Error, fmt.Sprintf("GitHub Contents API error: %s - %s", resp.Status, string(bodyBytes)), "")
		return nil, fmt.Errorf("GitHub API error: %s", resp.Status)
	}

	var content githubContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&content); err != nil {
		return nil, fmt.Errorf("failed to decode GitHub response: %w", err)
	}

	if content.Encoding != "base64" {
		return nil, fmt.Errorf("unexpected content encoding: %s", content.Encoding)
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(content.Content, "\n", ""))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 content: %w", err)
	}

	return decoded, nil
}
