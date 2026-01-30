package service

import (
	github_storage "github.com/raghavyuva/nixopus-api/internal/features/github-connector/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// IsOnboarded checks if a user is onboarded.
// A user is considered onboarded if they have:
// - Connected at least one GitHub connector (non-deleted)
func (s *UserService) IsOnboarded(userID string) (bool, error) {
	// Check for GitHub connectors
	githubConnectorStorage := &github_storage.GithubConnectorStorage{
		DB:  s.store.DB,
		Ctx: s.Ctx,
	}
	connectors, err := githubConnectorStorage.GetAllConnectors(userID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to check GitHub connectors for onboarding status", userID)
		return false, err
	}

	// If user has at least one non-deleted connector, they are onboarded
	if len(connectors) > 0 {
		return true, nil
	}

	// User has no GitHub connectors
	return false, nil
}
