package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// GitClient defines the interface for git operations
type GitClient interface {
	Clone(repoURL, destinationPath string) error
	Pull(repoURL, destinationPath string) error
	GetLatestCommitHash(repoURL string,accessToken string) (string, error)
}

// DefaultGitClient is the default implementation of GitClient
type DefaultGitClient struct {
	logger logger.Logger
}

// NewDefaultGitClient creates a new DefaultGitClient
func NewDefaultGitClient(logger logger.Logger) *DefaultGitClient {
	return &DefaultGitClient{
		logger: logger,
	}
}

// Clone clones a git repository to the specified path
func (g *DefaultGitClient) Clone(repoURL, destinationPath string) error {
	cmd := exec.Command("git", "clone", repoURL, destinationPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git clone failed: %s, output: %s", err.Error(), string(output))
	}
	return nil
}

func (g *DefaultGitClient) Pull(repoURL, destinationPath string) error {
	cmd := exec.Command("git", "pull", repoURL)
	cmd.Dir = destinationPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git pull failed: %s, output: %s", err.Error(), string(output))
	}
	return nil
}

func (g *DefaultGitClient) GetLatestCommitHash(repoURL string,accessToken string) (string, error) {
	parsedURL := strings.TrimSuffix(repoURL, ".git")
	urlParts := strings.Split(parsedURL, "/")
	if len(urlParts) < 2 {
		return "", fmt.Errorf("invalid repository URL format: %s", repoURL)
	}
	
	owner := urlParts[len(urlParts)-2]
	repo := urlParts[len(urlParts)-1]

	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/commits/HEAD", owner, repo)
	
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %s", err.Error())
	}
	
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
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
	
	g.logger.Log(logger.Info, fmt.Sprintf("Latest commit hash: %s", response.SHA), "")
	
	return response.SHA, nil
}