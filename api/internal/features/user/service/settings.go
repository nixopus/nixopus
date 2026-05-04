package service

import (
	"fmt"
	"time"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/types"
)

func (s *UserService) GetSettings(userID string) (*types.UserSettings, error) {
	ctx := fmt.Sprintf("user_id=%s", userID)
	s.logger.Log(logger.Info, "user service: GetSettings", ctx)
	return s.storage.GetUserSettings(userID)
}

func (s *UserService) UpdateFont(userID string, fontFamily string, fontSize int) (*types.UserSettings, error) {
	ctx := fmt.Sprintf("user_id=%s font_family=%s font_size=%d", userID, fontFamily, fontSize)
	s.logger.Log(logger.Info, "user service: UpdateFont", ctx)
	return s.storage.UpdateUserSettings(userID, map[string]interface{}{
		"font_family": fontFamily,
		"font_size":   fontSize,
		"updated_at":  time.Now(),
	})
}

func (s *UserService) UpdateTheme(userID string, theme string) (*types.UserSettings, error) {
	ctx := fmt.Sprintf("user_id=%s theme=%s", userID, theme)
	s.logger.Log(logger.Info, "user service: UpdateTheme", ctx)
	return s.storage.UpdateUserSettings(userID, map[string]interface{}{
		"theme":      theme,
		"updated_at": time.Now(),
	})
}

func (s *UserService) UpdateLanguage(userID string, language string) (*types.UserSettings, error) {
	ctx := fmt.Sprintf("user_id=%s language=%s", userID, language)
	s.logger.Log(logger.Info, "user service: UpdateLanguage", ctx)
	return s.storage.UpdateUserSettings(userID, map[string]interface{}{
		"language":   language,
		"updated_at": time.Now(),
	})
}

func (s *UserService) UpdateAutoUpdate(userID string, autoUpdate bool) (*types.UserSettings, error) {
	ctx := fmt.Sprintf("user_id=%s auto_update=%v", userID, autoUpdate)
	s.logger.Log(logger.Info, "user service: UpdateAutoUpdate", ctx)
	return s.storage.UpdateUserSettings(userID, map[string]interface{}{
		"auto_update": autoUpdate,
		"updated_at":  time.Now(),
	})
}

// GetUserPreferences retrieves user preferences
func (s *UserService) GetUserPreferences(userID string) (*types.UserPreferences, error) {
	ctx := fmt.Sprintf("user_id=%s", userID)
	s.logger.Log(logger.Info, "user service: GetUserPreferences", ctx)
	return s.storage.GetUserPreferences(userID)
}

// UpdateUserPreferences updates user preferences with the provided data
func (s *UserService) UpdateUserPreferences(userID string, preferences types.UserPreferencesData) (*types.UserPreferences, error) {
	ctx := fmt.Sprintf("user_id=%s", userID)
	s.logger.Log(logger.Info, "user service: UpdateUserPreferences", ctx)
	return s.storage.UpdateUserPreferences(userID, preferences)
}
