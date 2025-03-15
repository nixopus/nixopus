package service

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// createAuthenticatedRepoURL creates an authenticated URL for repository access
func (s *GithubConnectorService) CreateAuthenticatedRepoURL(repoURL, accessToken string) (string, error) {

	if strings.HasPrefix(repoURL, "https://") {
		parsedURL, err := url.Parse(repoURL)
		if err != nil {
			return "", fmt.Errorf("invalid repository URL: %w", err)
		}

		return fmt.Sprintf("https://oauth2:%s@github.com%s", accessToken, parsedURL.Path), nil

	} else if strings.HasPrefix(repoURL, "git@github.com") {
		parts := strings.Split(strings.TrimPrefix(repoURL, "git@github.com:"), "/")
		if len(parts) < 2 {
			return "", fmt.Errorf("invalid SSH repository URL format")
		}

		owner := parts[0]
		repo := strings.TrimSuffix(parts[len(parts)-1], ".git")

		return fmt.Sprintf("https://oauth2:%s@github.com/%s/%s.git", accessToken, owner, repo), nil
	}

	return "", fmt.Errorf("unsupported repository URL format")
}

// GetClonePath generates a path to clone a repository to.
//
// Parameters:
//
//	userID - the ID of the user whose repository to clone.
//	environment - the environment name to clone the repository to.
//	deploymentID - the ID of the deployment.
//	latestCommitHash - the hash of the latest commit.
//
// Returns:
//
//	string - the path to which to clone the repository.
//	string - the context information path.
//	bool - whether to pull the repository instead of cloning.
func (s *GithubConnectorService) GetClonePath(userID, environment string) (string, bool, error) {
	repoBaseURL := os.Getenv("MOUNT_PATH")
	clonePath := filepath.Join(repoBaseURL, userID, environment)
	var shouldPull bool

	if _, err := os.Stat(clonePath); err == nil {
		shouldPull = true
	}

	if err := os.MkdirAll(clonePath, 0755); err != nil {
		return "", false, err
	}

	return clonePath, shouldPull, nil
}
