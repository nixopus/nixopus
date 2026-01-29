package service

import (
	"context"
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (s *GithubConnectorService) handleGitPull(ctx context.Context, authenticatedURL, clonePath string, userID string) error {
	gitClient, err := s.getGitClient(ctx)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to get git client: %s", err.Error()), userID)
		return err
	}

	hasChanges, err := gitClient.HasUncommittedChanges(clonePath)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to check for uncommitted changes: %s", err.Error()), userID)
		return err
	}

	if hasChanges {
		s.logger.Log(logger.Info, "Discarding local changes for clean state", userID)
		if err := gitClient.ResetHard(clonePath); err != nil {
			s.logger.Log(logger.Error, fmt.Sprintf("Failed to reset repository: %s", err.Error()), userID)
			return err
		}
	}

	s.logger.Log(logger.Info, "Pulling latest changes", userID)
	if err := gitClient.Pull(authenticatedURL, clonePath); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to pull repository: %s", err.Error()), userID)
		return err
	}

	return nil
}
