package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
)

func (s *NotificationService) UpdatePreference(preferences notification.UpdatePreferenceRequest, userID uuid.UUID) error {
	s.logger.Log(logger.Info, fmt.Sprintf("Updating preference: Category=%s, Type=%s, Enabled=%v",
		preferences.Category, preferences.Type, preferences.Enabled), "")
	s.storage.UpdatePreference(s.Ctx, preferences, userID)
	return nil
}
