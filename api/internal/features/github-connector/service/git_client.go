package service

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// GitClient defines the interface for git operations
type GitClient interface {
	Clone(repoURL, destinationPath string) error
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

func extractRepoName(repoURL string) string {
	if strings.HasPrefix(repoURL, "https://") {
		parts := strings.Split(repoURL, "/")
		repoName := parts[len(parts)-1]
		repoName = strings.TrimSuffix(repoName, ".git")
		return repoName
	}

	if strings.HasPrefix(repoURL, "git@") {
		parts := strings.Split(repoURL, ":")
		if len(parts) >= 2 {
			repoPath := parts[1]
			pathParts := strings.Split(repoPath, "/")
			repoName := pathParts[len(pathParts)-1]
			repoName = strings.TrimSuffix(repoName, ".git")
			return repoName
		}
	}

	sanitized := strings.ReplaceAll(repoURL, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, ":", "_")
	sanitized = strings.ReplaceAll(sanitized, ".", "_")
	return sanitized
}
