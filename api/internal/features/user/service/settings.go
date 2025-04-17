package service

import (
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
)


func (s *UserService) GetSettings(userID string) (*types.UserSettings, error) {
	s.logger.Log(logger.Info, "getting user settings", "")
	return s.storage.GetUserSettings(userID)
}

func (s *UserService) UpdateFont(userID string, fontFamily string, fontSize int) (*types.UserSettings, error) {
	s.logger.Log(logger.Info, "updating user font settings", "")
	return s.storage.UpdateUserSettings(userID, map[string]interface{}{
		"font_family": fontFamily,
		"font_size":   fontSize,
		"updated_at":  time.Now(),
	})
}

func (s *UserService) UpdateTheme(userID string, theme string) (*types.UserSettings, error) {
	s.logger.Log(logger.Info, "updating user theme", "")
	return s.storage.UpdateUserSettings(userID, map[string]interface{}{
		"theme":      theme,
		"updated_at": time.Now(),
	})
}

func (s *UserService) UpdateLanguage(userID string, language string) (*types.UserSettings, error) {
	s.logger.Log(logger.Info, "updating user language", "")
	return s.storage.UpdateUserSettings(userID, map[string]interface{}{
		"language":   language,
		"updated_at": time.Now(),
	})
}

func (s *UserService) UpdateAutoUpdate(userID string, autoUpdate bool) (*types.UserSettings, error) {
	s.logger.Log(logger.Info, "updating user auto update setting", "")
	return s.storage.UpdateUserSettings(userID, map[string]interface{}{
		"auto_update": autoUpdate,
		"updated_at":  time.Now(),
	})
}
