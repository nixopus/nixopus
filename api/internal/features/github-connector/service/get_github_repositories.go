package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetGithubRepositoriesPaginated fetches repositories for the user's GitHub installation with pagination.
// If connectorID is provided, it uses that specific connector. Otherwise, it finds a connector with a valid installation_id.
// If search is provided, it fetches all repositories and filters them by the search term before applying pagination.
func (c *GithubConnectorService) GetGithubRepositoriesPaginated(userID string, page int, pageSize int, connectorID string, search string, sortBy string, sortDirection string) ([]shared_types.GithubRepository, int, error) {
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
		for i := range connectors {
			if connectors[i].InstallationID != "" && connectors[i].InstallationID != " " {
				connectorToUse = &connectors[i]
				break
			}
		}
		if connectorToUse == nil {
			c.logger.Log(logger.Error, "No connector with valid installation_id found for user", userID)
			return nil, 0, fmt.Errorf("no connector with valid installation found")
		}
	}

	if connectorToUse.InstallationID == "" || connectorToUse.InstallationID == " " {
		c.logger.Log(logger.Error, fmt.Sprintf("Connector %s has empty installation_id", connectorToUse.ID.String()), userID)
		return nil, 0, fmt.Errorf("connector has no installation_id")
	}

	installation_id := connectorToUse.InstallationID
	jwt := GenerateJwt(connectorToUse)

	if jwt == "" {
		c.logger.Log(logger.Error, "Failed to generate app JWT", "")
		return nil, 0, fmt.Errorf("failed to generate app JWT: GitHub App credentials are not configured")
	}

	accessToken, err := c.getInstallationToken(jwt, installation_id)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("Failed to get installation token: %s", err.Error()), "")
		if strings.Contains(err.Error(), "installation not found") {
			return nil, 0, fmt.Errorf("invalid GitHub installation: %w. Please reconnect your GitHub account", err)
		}
		return nil, 0, err
	}

	var repositories []shared_types.GithubRepository
	var totalCount int

	if search != "" {
		repositories, totalCount, err = c.fetchAllAndFilter(accessToken, page, pageSize, search, sortBy, sortDirection)
		if err != nil {
			return nil, 0, err
		}
	} else if sortBy != "" {
		repositories, totalCount, err = c.fetchAllSortAndPaginate(accessToken, page, pageSize, sortBy, sortDirection)
		if err != nil {
			return nil, 0, err
		}
	} else {
		repositories, totalCount, err = c.fetchPaginatedRepositories(accessToken, page, pageSize)
		if err != nil {
			return nil, 0, err
		}
	}

	return repositories, totalCount, nil
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

// fetchAllAndFilter fetches all repositories from GitHub, filters by search term, sorts, and returns paginated results
func (c *GithubConnectorService) fetchAllAndFilter(accessToken string, page int, pageSize int, search string, sortBy string, sortDirection string) ([]shared_types.GithubRepository, int, error) {
	allRepos := []shared_types.GithubRepository{}
	currentPage := 1
	perPage := 100

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

		if len(allRepos) >= response.TotalCount || len(response.Repositories) < perPage {
			break
		}

		currentPage++
	}

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

	if sortBy != "" {
		filteredRepos = c.sortRepositories(filteredRepos, sortBy, sortDirection)
	}

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

// sortRepositories sorts repositories based on the provided sort field and direction
func (c *GithubConnectorService) sortRepositories(repos []shared_types.GithubRepository, sortBy string, sortDirection string) []shared_types.GithubRepository {
	if len(repos) == 0 {
		return repos
	}

	if sortDirection == "" {
		sortDirection = "asc"
	}

	sorted := make([]shared_types.GithubRepository, len(repos))
	copy(sorted, repos)

	sort.Slice(sorted, func(i, j int) bool {
		var comparison bool

		switch sortBy {
		case "name":
			comparison = strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
		case "stargazers_count", "stars":
			comparison = sorted[i].StargazersCount < sorted[j].StargazersCount
		default:
			comparison = strings.ToLower(sorted[i].Name) < strings.ToLower(sorted[j].Name)
		}

		if sortDirection == "desc" {
			return !comparison
		}
		return comparison
	})

	return sorted
}

// fetchAllSortAndPaginate fetches all repositories from GitHub, sorts them, and returns paginated results
func (c *GithubConnectorService) fetchAllSortAndPaginate(accessToken string, page int, pageSize int, sortBy string, sortDirection string) ([]shared_types.GithubRepository, int, error) {
	allRepos := []shared_types.GithubRepository{}
	currentPage := 1
	perPage := 100

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

		if len(allRepos) >= response.TotalCount || len(response.Repositories) < perPage {
			break
		}

		currentPage++
	}

	allRepos = c.sortRepositories(allRepos, sortBy, sortDirection)

	totalCount := len(allRepos)
	start := (page - 1) * pageSize
	if start > totalCount {
		start = totalCount
	}
	end := start + pageSize
	if end > totalCount {
		end = totalCount
	}

	return allRepos[start:end], totalCount, nil
}
