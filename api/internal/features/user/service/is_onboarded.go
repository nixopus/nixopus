package service

import (
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
)

// IsOnboarded checks if a user is onboarded by reading the is_onboarded field from the database.
func (s *UserService) IsOnboarded(userID string) (bool, error) {
	logData := fmt.Sprintf("user_id=%s", userID)
	isOnboarded, err := s.storage.GetIsOnboarded(userID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("user service: IsOnboarded: %v", err), logData)
		return false, err
	}

	return isOnboarded, nil
}
