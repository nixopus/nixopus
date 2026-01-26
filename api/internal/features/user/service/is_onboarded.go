package service

import (
	"github.com/google/uuid"
	auth_storage "github.com/raghavyuva/nixopus-api/internal/features/auth/storage"
	github_storage "github.com/raghavyuva/nixopus-api/internal/features/github-connector/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// IsOnboarded checks if a user is onboarded.
// A user is considered onboarded if they have either:
// - Generated at least one API key (non-revoked)
// - Connected at least one GitHub connector (non-deleted)
func (s *UserService) IsOnboarded(userID string) (bool, error) {
	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Log(logger.Error, "Invalid user ID format", userID)
		return false, err
	}

	// Check for API keys
	apiKeyStorage := &auth_storage.APIKeyStorage{
		DB:  s.store.DB,
		Ctx: s.Ctx,
	}
	apiKeys, err := apiKeyStorage.FindAPIKeysByUserID(parsedUserID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to check API keys for onboarding status", userID)
		return false, err
	}

	// If user has at least one non-revoked API key, they are onboarded
	if len(apiKeys) > 0 {
		return true, nil
	}

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

	// User has neither API keys nor GitHub connectors
	return false, nil
}
