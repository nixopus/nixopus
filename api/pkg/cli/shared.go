package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/config"
)

// IsGitRepo checks if the given path is a git repository
func IsGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// getGitBranch gets the current git branch
// we will make sure that the current branch is git enabled
func getGitBranch() string {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// getGitInfo gets the git repository name and remote URL
func getGitInfo() (string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	repoName := filepath.Base(cwd)

	// Try to get git remote URL
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return repoName, repoName, nil
	}

	repoURL := strings.TrimSpace(string(output))
	if repoURL == "" {
		return repoName, repoName, nil
	}

	// Extract repo name from URL (e.g., github.com/user/repo.git -> repo)
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		lastPart := strings.TrimSuffix(parts[len(parts)-1], ".git")
		if lastPart != "" {
			repoName = lastPart
		}
	}

	return repoName, repoURL, nil
}

// BuildDomainURL builds the domain URL from project ID using config.
// Format: https://{first-8-chars-of-project-id}.{deploy_domain}
func BuildDomainURL(projectID string) string {
	return config.BuildDeployDomainURL(projectID)
}

// RemoveFromSlice removes an item from a string slice
func RemoveFromSlice(slice []string, item string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if s != item {
			result = append(result, s)
		}
	}
	return result
}
