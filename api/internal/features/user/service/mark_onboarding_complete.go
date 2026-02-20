package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// MarkOnboardingComplete marks a user's onboarding as complete by setting is_onboarded to true.
func (s *UserService) MarkOnboardingComplete(userID string) error {
	err := s.storage.MarkOnboardingComplete(userID)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to mark onboarding as complete", userID)
		return err
	}

	s.logger.Log(logger.Info, "Onboarding marked as complete", userID)
	return nil
}
