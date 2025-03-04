package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
)

// UpdateSmtp updates an existing SMTP configuration in the database.
//
// It takes a notification.UpdateSMTPConfigRequest as a parameter,
// logs an info message to the logger, and calls the storage layer
// to perform the update operation. It returns an error if the
// storage operation fails.
func (s *NotificationService) UpdateSmtp(SMTPConfigs notification.UpdateSMTPConfigRequest) error {
	s.logger.Log(logger.Info, "Updating SMTP configuration", "")
	return s.storage.UpdateSmtp(&SMTPConfigs)
}
