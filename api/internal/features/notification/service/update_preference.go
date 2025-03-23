package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
)

// UpdatePreference updates a user's notification preference.
//
// It will update the corresponding preference item in the database.
//
// The function validates input parameters and returns an error if validation fails
// or if there's an error updating the preference in storage.
//
// The function will log an info message with the details of the preference update.
func (s *NotificationService) UpdatePreference(preferences notification.UpdatePreferenceRequest, userID uuid.UUID) error {
	if preferences.Category == "" {
		return fmt.Errorf("category cannot be empty")
	}

	if preferences.Type == "" {
		return fmt.Errorf("type cannot be empty")
	}

	if userID == uuid.Nil {
		return fmt.Errorf("userID cannot be empty")
	}

	s.logger.Log(logger.Info, fmt.Sprintf("Updating preference: Category=%s, Type=%s, Enabled=%v",
		preferences.Category, preferences.Type, preferences.Enabled), "")

	err := s.storage.UpdatePreference(s.Ctx, preferences, userID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to update preference: %v", err), "")
		return fmt.Errorf("failed to update preference: %w", err)
	}

	return nil
}
