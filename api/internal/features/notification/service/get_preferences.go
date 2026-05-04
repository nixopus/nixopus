package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
)

// GetPreferences fetches the notification preferences for the given user
//
// If the user has not set any preferences before, this will return an empty
// response. If the user has no preferences, this will return an error.
func (s *NotificationService) GetPreferences(userID uuid.UUID) (*notification.GetPreferencesResponse, error) {
	s.logger.Log(logger.Info, "notification service: GetPreferences", fmt.Sprintf("user_id=%s", userID))
	return s.storage.GetPreferences(s.Ctx, userID)
}
