package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// IsOnboarded checks if a user is onboarded by reading the is_onboarded field from the database.
func (s *UserService) IsOnboarded(userID string) (bool, error) {
	isOnboarded, err := s.storage.GetIsOnboarded(userID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get onboarding status", userID)
		return false, err
	}

	return isOnboarded, nil
}
