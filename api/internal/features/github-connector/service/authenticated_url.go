package service

import (
	"context"
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
//
// Parameters:
//
//	userID - the ID of the user whose repository to clone.
//	environment - the environment name to clone the repository to.
//
// Returns:
//
//	string - the path to which to clone the repository.
//	bool - whether to pull the repository instead of cloning.
//	error - any error that occurred.
func (s *GithubConnectorService) GetClonePath(ctx context.Context, userID, environment, applicationID string) (string, bool, error) {
	repoBaseURL := config.AppConfig.Deployment.MountPath
	clonePath := filepath.Join(repoBaseURL, userID, environment, applicationID)
	var shouldPull bool

	sshManager, err := s.getSSHManager(ctx)
	if err != nil {
		return "", false, fmt.Errorf("failed to get SSH manager: %w", err)
	}
	client, err := sshManager.Connect()
	if err != nil {
		return "", false, fmt.Errorf("failed to connect via SSH: %w", err)
	}
	defer client.Close()

	sftp, err := client.NewSftp()
	if err != nil {
		return "", false, fmt.Errorf("failed to create SFTP client: %w", err)
	}
	defer sftp.Close()

	info, err := sftp.Stat(clonePath)
	if err == nil && info.IsDir() {
		shouldPull = true
	}

	if !shouldPull {
		err = sftp.MkdirAll(clonePath)
		if err != nil {
			return "", false, fmt.Errorf("failed to create directory via SFTP: %w", err)
		}
	}

	return clonePath, shouldPull, nil
}
