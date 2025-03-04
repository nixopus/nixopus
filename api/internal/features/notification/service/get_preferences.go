package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
)

// GetPreferences fetches the notification preferences for the given user
//
// If the user has not set any preferences before, this will return an empty
// response. If the user has no preferences, this will return an error.
func (s *NotificationService) GetPreferences(userID uuid.UUID) (*notification.GetPreferencesResponse, error) {
	return s.storage.GetPreferences(s.Ctx, userID)
}
