package service

import (
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
)

// UpdateSmtp updates the SMTP configuration.
//
// It validates the input parameters and updates the SMTP configuration in storage.
// Returns an error if validation fails or if there's an error updating the configuration.
func (s *NotificationService) UpdateSmtp(config notification.UpdateSMTPConfigRequest) error {
	if s.storage == nil {
		return fmt.Errorf("storage layer not initialized")
	}

	if config.Host == nil {
		return fmt.Errorf("SMTP host cannot be empty")
	}

	s.logger.Log(logger.Info, fmt.Sprintf("Updating SMTP configuration: Host=%s", *config.Host), "")

	err := s.storage.UpdateSmtp(&config)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to update SMTP configuration: %v", err), "")
		return fmt.Errorf("failed to update SMTP configuration: %w", err)
	}

	return nil
}
