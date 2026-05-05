package service

import (
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
)

// MarkOnboardingComplete marks a user's onboarding as complete by setting is_onboarded to true.
func (s *UserService) MarkOnboardingComplete(userID string) error {
	logData := fmt.Sprintf("user_id=%s", userID)
	err := s.storage.MarkOnboardingComplete(userID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("user service: MarkOnboardingComplete: %v", err), logData)
		return err
	}

	s.logger.Log(logger.Info, "user service: MarkOnboardingComplete ok", logData)
	return nil
}
