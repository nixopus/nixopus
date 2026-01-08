package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetGithubRepositoriesPaginated fetches repositories for the user's GitHub installation with pagination.
// If connectorID is provided, it uses that specific connector. Otherwise, it finds a connector with a valid installation_id.
// If search is provided, it fetches all repositories and filters them by the search term before applying pagination.
func (c *GithubConnectorService) GetGithubRepositoriesPaginated(userID string, page int, pageSize int, connectorID string, search string) ([]shared_types.GithubRepository, int, error) {
	accessToken, err := c.getAccessTokenForUser(userID, connectorID)
	if err != nil {
		return nil, 0, err
	}

	// If search is provided, fetch all repositories and filter
	if search != "" {
		return c.fetchAllAndFilter(accessToken, page, pageSize, search)
	}

	// No search - use direct pagination from GitHub API
	return c.fetchPaginatedRepositories(accessToken, page, pageSize)
}

// fetchPaginatedRepositories fetches a single page of repositories from GitHub
func (c *GithubConnectorService) fetchPaginatedRepositories(accessToken string, page int, pageSize int) ([]shared_types.GithubRepository, int, error) {
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

// fetchAllAndFilter fetches all repositories from GitHub, filters by search term, and returns paginated results
func (c *GithubConnectorService) fetchAllAndFilter(accessToken string, page int, pageSize int, search string) ([]shared_types.GithubRepository, int, error) {
	allRepos := []shared_types.GithubRepository{}
	currentPage := 1
	perPage := 100 // Max allowed by GitHub API

	client := &http.Client{}

	for {
		url := fmt.Sprintf("https://api.github.com/installation/repositories?per_page=%d&page=%d", perPage, currentPage)
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

		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			c.logger.Log(logger.Error, fmt.Sprintf("GitHub API error: %s - %s", resp.Status, string(bodyBytes)), "")
			return nil, 0, fmt.Errorf("GitHub API error: %s", resp.Status)
		}

		var response struct {
			TotalCount   int                             `json:"total_count"`
			Repositories []shared_types.GithubRepository `json:"repositories"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			c.logger.Log(logger.Error, err.Error(), "")
			return nil, 0, err
		}
		resp.Body.Close()

		allRepos = append(allRepos, response.Repositories...)

		// Check if we've fetched all repositories
		if len(allRepos) >= response.TotalCount || len(response.Repositories) < perPage {
			break
		}

		currentPage++
	}

	// Filter repositories by search term (case-insensitive search on name and description)
	searchLower := strings.ToLower(search)
	filteredRepos := []shared_types.GithubRepository{}
	for _, repo := range allRepos {
		nameLower := strings.ToLower(repo.Name)
		descLower := ""
		if repo.Description != nil {
			descLower = strings.ToLower(*repo.Description)
		}
		if strings.Contains(nameLower, searchLower) || strings.Contains(descLower, searchLower) {
			filteredRepos = append(filteredRepos, repo)
		}
	}

	// Apply pagination to filtered results
	totalCount := len(filteredRepos)
	start := (page - 1) * pageSize
	if start > totalCount {
		start = totalCount
	}
	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}

	return filteredRepos[start:end], totalCount, nil
}
