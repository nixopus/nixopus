package gh

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// GetRepositoriesPaginated lists installation repos with optional search/sort.
func (a *API) GetRepositoriesPaginated(userID string, page, pageSize int, connectorID, search, sortBy, sortDirection string) ([]shared_types.GithubRepository, int, error) {
	connectors, err := a.Storage.GetAllConnectors(userID)
	ctxBase := fmt.Sprintf("user_id=%s", userID)
	if err != nil {
		a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: GetRepositoriesPaginated list connectors: %v", err), ctxBase)
		return nil, 0, err
	}
	if len(connectors) == 0 {
		a.Logger.Log(logger.Debug, "github connector service: GetRepositoriesPaginated no connectors", ctxBase)
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
			a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: GetRepositoriesPaginated connector_id=%s not found", connectorID), ctxBase)
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
			a.Logger.Log(logger.Error, "github connector service: GetRepositoriesPaginated no connector with installation_id", ctxBase)
			return nil, 0, fmt.Errorf("no connector with valid installation found")
		}
	}
	if connectorToUse.InstallationID == "" || connectorToUse.InstallationID == " " {
		a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: GetRepositoriesPaginated empty installation_id connector_id=%s", connectorToUse.ID.String()), ctxBase)
		return nil, 0, fmt.Errorf("connector has no installation_id")
	}
	installationID := connectorToUse.InstallationID
	jwt := GenerateJwt(connectorToUse)
	if jwt == "" {
		a.Logger.Log(logger.Error, "github connector service: GetRepositoriesPaginated JWT generation failed", ctxBase)
		return nil, 0, fmt.Errorf("failed to generate app JWT: GitHub App credentials are not configured")
	}
	accessToken, err := InstallationToken(jwt, installationID)
	if err != nil {
		a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: GetRepositoriesPaginated installation token: %v", err), ctxBase)
		if strings.Contains(err.Error(), "installation not found") {
			return nil, 0, fmt.Errorf("invalid GitHub installation: %w. Please reconnect your GitHub account", err)
		}
		return nil, 0, err
	}
	var repositories []shared_types.GithubRepository
	var totalCount int
	if search != "" {
		repositories, totalCount, err = a.fetchAllAndFilter(accessToken, page, pageSize, search, sortBy, sortDirection)
	} else if sortBy != "" {
		repositories, totalCount, err = a.fetchAllSortAndPaginate(accessToken, page, pageSize, sortBy, sortDirection)
	} else {
		repositories, totalCount, err = a.fetchPaginatedRepositories(accessToken, page, pageSize)
	}
	if err != nil {
		return nil, 0, err
	}
	return repositories, totalCount, nil
}

func (a *API) fetchPaginatedRepositories(accessToken string, page, pageSize int) ([]shared_types.GithubRepository, int, error) {
	client := &http.Client{}
	u := fmt.Sprintf("%s/installation/repositories?per_page=%d&page=%d", APIBaseURL, pageSize, page)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: installation repos: %v", err), "")
		return nil, 0, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nixopus")
	resp, err := client.Do(req)
	if err != nil {
		a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: installation repos: %v", err), "")
		return nil, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: GitHub API error: %s - %s", resp.Status, string(bodyBytes)), "")
		return nil, 0, fmt.Errorf("GitHub API error: %s", resp.Status)
	}
	var response struct {
		TotalCount   int                             `json:"total_count"`
		Repositories []shared_types.GithubRepository `json:"repositories"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: installation repos: %v", err), "")
		return nil, 0, err
	}
	return response.Repositories, response.TotalCount, nil
}

func (a *API) fetchAllAndFilter(accessToken string, page, pageSize int, search, sortBy, sortDirection string) ([]shared_types.GithubRepository, int, error) {
	allRepos := []shared_types.GithubRepository{}
	currentPage := 1
	perPage := 100
	client := &http.Client{}
	for {
		u := fmt.Sprintf("%s/installation/repositories?per_page=%d&page=%d", APIBaseURL, perPage, currentPage)
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: installation repos: %v", err), "")
			return nil, 0, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "nixopus")
		resp, err := client.Do(req)
		if err != nil {
			a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: installation repos: %v", err), "")
			return nil, 0, err
		}
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: GitHub API error: %s - %s", resp.Status, string(bodyBytes)), "")
			return nil, 0, fmt.Errorf("GitHub API error: %s", resp.Status)
		}
		var response struct {
			TotalCount   int                             `json:"total_count"`
			Repositories []shared_types.GithubRepository `json:"repositories"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: installation repos: %v", err), "")
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
		filteredRepos = a.sortRepositories(filteredRepos, sortBy, sortDirection)
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

func (a *API) sortRepositories(repos []shared_types.GithubRepository, sortBy, sortDirection string) []shared_types.GithubRepository {
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

func (a *API) fetchAllSortAndPaginate(accessToken string, page, pageSize int, sortBy, sortDirection string) ([]shared_types.GithubRepository, int, error) {
	allRepos := []shared_types.GithubRepository{}
	currentPage := 1
	perPage := 100
	client := &http.Client{}
	for {
		u := fmt.Sprintf("%s/installation/repositories?per_page=%d&page=%d", APIBaseURL, perPage, currentPage)
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: installation repos: %v", err), "")
			return nil, 0, err
		}
		req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
		req.Header.Set("Accept", "application/vnd.github.v3+json")
		req.Header.Set("User-Agent", "nixopus")
		resp, err := client.Do(req)
		if err != nil {
			a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: installation repos: %v", err), "")
			return nil, 0, err
		}
		if resp.StatusCode != http.StatusOK {
			bodyBytes, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: GitHub API error: %s - %s", resp.Status, string(bodyBytes)), "")
			return nil, 0, fmt.Errorf("GitHub API error: %s", resp.Status)
		}
		var response struct {
			TotalCount   int                             `json:"total_count"`
			Repositories []shared_types.GithubRepository `json:"repositories"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			resp.Body.Close()
			a.Logger.Log(logger.Error, fmt.Sprintf("github connector service: installation repos: %v", err), "")
			return nil, 0, err
		}
		resp.Body.Close()
		allRepos = append(allRepos, response.Repositories...)
		if len(allRepos) >= response.TotalCount || len(response.Repositories) < perPage {
			break
		}
		currentPage++
	}
	allRepos = a.sortRepositories(allRepos, sortBy, sortDirection)
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
