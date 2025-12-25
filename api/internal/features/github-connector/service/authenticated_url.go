package service

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/config"
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
// Always creates a fresh directory for cloning (repositories are cleaned up after use).
//
// Parameters:
//
//	userID - the ID of the user whose repository to clone.
//	environment - the environment name to clone the repository to.
//	applicationID - the ID of the application to clone the repository for.
//
// Returns:
//
//	string - the path to which to clone the repository.
//	error - any error that occurred.
func (s *GithubConnectorService) GetClonePath(userID, environment, applicationID string) (string, error) {
	repoBaseURL := config.AppConfig.Deployment.MountPath
	clonePath := filepath.Join(repoBaseURL, userID, environment, applicationID)

	client, err := s.ssh.Connect()
	if err != nil {
		return "", fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer client.Close()

	sftp, err := client.NewSftp()
	if err != nil {
		return "", fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftp.Close()

	// Always ensure the directory exists (will be cleaned up after build)
	err = sftp.MkdirAll(clonePath)
	if err != nil {
		return "", fmt.Errorf("failed to create directory via SFTP: %w", err)
	}

	return clonePath, nil
}
