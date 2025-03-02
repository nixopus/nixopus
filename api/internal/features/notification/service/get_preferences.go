package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
)

func (s *NotificationService) GetPreferences(userID uuid.UUID) (*notification.GetPreferencesResponse, error) {
	return s.storage.GetPreferences(s.Ctx, userID)
}
