package service

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

var repoBaseURL = "/Users/shaale/nixopus-configs"

// createAuthenticatedRepoURL creates an authenticated URL for repository access
func (s *GithubConnectorService) createAuthenticatedRepoURL(repoURL, accessToken string) (string, error) {

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

// getClonePath generates a path to clone a repository to.
//
// Parameters:
//
//	userID - the ID of the user whose repository to clone.
//	environment - the environment name to clone the repository to.
//	repoURL - the URL of the repository to clone.
//
// Returns:
//
//	string - the path to which to clone the repository.
func (s *GithubConnectorService) getClonePath(userID, environment, repoURL string) string {
	repoName := extractRepoName(repoURL)

	clonePath := filepath.Join(repoBaseURL, userID, environment, repoName)

	os.MkdirAll(filepath.Dir(clonePath), 0755)

	return clonePath
}
