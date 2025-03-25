package service

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
)

// GitClient defines the interface for git operations
type GitClient interface {
	Clone(repoURL, destinationPath string) error
	Pull(repoURL, destinationPath string) error
	GetLatestCommitHash(repoURL string, accessToken string) (string, error)
	SetHeadToCommitHash(repoURL, destinationPath, commitHash string) error
}

// DefaultGitClient is the default implementation of GitClient
type DefaultGitClient struct {
	logger logger.Logger
	ssh    *ssh.SSH
}

// NewDefaultGitClient creates a new DefaultGitClient
func NewDefaultGitClient(logger logger.Logger, ssh *ssh.SSH) *DefaultGitClient {
	return &DefaultGitClient{
		logger: logger,
		ssh:    ssh,
	}
}

// Clone clones a git repository to the specified path
func (g *DefaultGitClient) Clone(repoURL, destinationPath string) error {
	client, err := g.ssh.ConnectWithPassword()
	if err != nil {
		return fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer client.Close()

	cmd := fmt.Sprintf("git clone %s %s", repoURL, destinationPath)
	output, err := client.Run(cmd)
	if err != nil {
		return fmt.Errorf("git clone failed: %s, output: %s", err.Error(),output)
	}

	g.logger.Log(logger.Info, fmt.Sprintf("Successfully cloned repository to %s", destinationPath), "")
	return nil
}

// Pull updates a git repository from remote
func (g *DefaultGitClient) Pull(repoURL, destinationPath string) error {
	client, err := g.ssh.ConnectWithPassword()
	if err != nil {
		return fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer client.Close()

	cmd := fmt.Sprintf("cd %s && git pull %s", destinationPath, repoURL)
	output, err := client.Run(cmd)
	if err != nil {
		return fmt.Errorf("git pull failed: %s, output: %s", err.Error(),output)
	}

	g.logger.Log(logger.Info, fmt.Sprintf("Successfully pulled latest changes for repository at %s", destinationPath), "")
	return nil
}

// GetLatestCommitHash retrieves the latest commit hash from the repository
func (g *DefaultGitClient) GetLatestCommitHash(repoURL string, accessToken string) (string, error) {
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

// SetHeadToCommitHash sets the HEAD of the repository to a specific commit hash
func (g *DefaultGitClient) SetHeadToCommitHash(repoURL, destinationPath, commitHash string) error {
	// Connect to SSH
	client, err := g.ssh.ConnectWithPassword()
	if err != nil {
		return fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer client.Close()

	cmd := fmt.Sprintf("cd %s && git checkout %s", destinationPath, commitHash)
	output, err := client.Run(cmd)
	if err != nil {
		return fmt.Errorf("git checkout failed: %s, output: %s", err.Error(),output)
	}

	g.logger.Log(logger.Info, fmt.Sprintf("Successfully checked out commit %s at %s", commitHash, destinationPath), "")
	return nil
}