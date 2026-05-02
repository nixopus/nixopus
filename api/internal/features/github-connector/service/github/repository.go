package gh

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// GetRepositoryByID fetches repository metadata from GitHub by numeric repo ID.
func (a *API) GetRepositoryByID(userID string, repoID uint64) (*shared_types.GithubRepository, error) {
	connectors, err := a.Storage.GetAllConnectors(userID)
	if err != nil {
		a.Logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	if len(connectors) == 0 {
		a.Logger.Log(logger.Error, "No connectors found for user", userID)
		return nil, fmt.Errorf("no connectors found for user")
	}
	jwt := GenerateJwt(&connectors[0])
	if jwt == "" {
		return nil, fmt.Errorf("failed to generate app JWT")
	}
	accessToken, err := InstallationToken(jwt, connectors[0].InstallationID)
	if err != nil {
		a.Logger.Log(logger.Error, fmt.Sprintf("Failed to get installation token: %s", err.Error()), "")
		return nil, err
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/repositories/%d", APIBaseURL, repoID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nixopus")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(bodyBytes))
	}
	var repo shared_types.GithubRepository
	if err := json.NewDecoder(resp.Body).Decode(&repo); err != nil {
		return nil, err
	}
	return &repo, nil
}

// GetRepositoryBranches lists branches for a repo (owner/repo name or numeric ID).
func (a *API) GetRepositoryBranches(userID, repositoryName string) ([]shared_types.GithubRepositoryBranch, error) {
	connectors, err := a.Storage.GetAllConnectors(userID)
	if err != nil {
		a.Logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	if len(connectors) == 0 {
		return []shared_types.GithubRepositoryBranch{}, nil
	}
	jwt := GenerateJwt(&connectors[0])
	if jwt == "" {
		return nil, fmt.Errorf("failed to generate app JWT")
	}
	accessToken, err := InstallationToken(jwt, connectors[0].InstallationID)
	if err != nil {
		a.Logger.Log(logger.Error, fmt.Sprintf("Failed to get installation token: %s", err.Error()), "")
		return nil, err
	}
	var repoFullName string
	if repoID, err := strconv.ParseUint(repositoryName, 10, 64); err == nil {
		repo, err := a.GetRepositoryByID(userID, repoID)
		if err != nil {
			return nil, fmt.Errorf("failed to get repository details: %s", err.Error())
		}
		repoFullName = repo.FullName
	} else {
		repoFullName = repositoryName
	}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/repos/%s/branches", APIBaseURL, repoFullName), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "nixopus")
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(bodyBytes))
	}
	var branches []shared_types.GithubRepositoryBranch
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, err
	}
	return branches, nil
}

type contentResponse struct {
	Content  string `json:"content"`
	Encoding string `json:"encoding"`
}

// GetRepositoryFileContent fetches file bytes from the GitHub Contents API.
func (a *API) GetRepositoryFileContent(userID, repository, branch, filePath string) ([]byte, error) {
	connectors, err := a.Storage.GetAllConnectors(userID)
	if err != nil {
		a.Logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}
	if len(connectors) == 0 {
		return nil, fmt.Errorf("no GitHub connectors found for user")
	}
	jwt := GenerateJwt(&connectors[0])
	if jwt == "" {
		return nil, fmt.Errorf("failed to generate GitHub App JWT")
	}
	accessToken, err := InstallationToken(jwt, connectors[0].InstallationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get installation token: %w", err)
	}
	var repoFullName string
	if repoID, parseErr := strconv.ParseUint(repository, 10, 64); parseErr == nil {
		repo, err := a.GetRepositoryByID(userID, repoID)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve repository ID %d: %w", repoID, err)
		}
		repoFullName = repo.FullName
	} else {
		repoFullName = repository
	}
	cleanPath := strings.TrimPrefix(filePath, "/")
	apiURL := fmt.Sprintf("%s/repos/%s/contents/%s?ref=%s",
		APIBaseURL, repoFullName, cleanPath, url.QueryEscape(branch))
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
		return nil, fmt.Errorf("GitHub API error: %s - %s", resp.Status, string(bodyBytes))
	}
	var content contentResponse
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

// LatestCommitHash returns the SHA of HEAD for a GitHub repo using the REST API.
func LatestCommitHash(log logger.Logger, repoURL string, accessToken string) (string, error) {
	parsedURL := strings.TrimSuffix(repoURL, ".git")
	parts := strings.Split(parsedURL, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("invalid repository URL format: %s", repoURL)
	}
	owner := parts[len(parts)-2]
	repo := parts[len(parts)-1]
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/repos/%s/%s/commits/HEAD", APIBaseURL, owner, repo), nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %s", err.Error())
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	resp, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %s", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	var response struct {
		SHA string `json:"sha"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to decode response: %s", err.Error())
	}
	log.Log(logger.Info, fmt.Sprintf("Latest commit hash: %s", response.SHA), "")
	return response.SHA, nil
}
