package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
)

// AddSmtp adds a new SMTP configuration to the database.
//
// It takes a notification.CreateSMTPConfigRequest and a userID as parameters.
// It logs an info message to the logger.
// It calls notification.NewSMTPConfig to create a new shared_types.SMTPConfigs with the given request and userID.
// It calls s.storage.AddSmtp with the new config.
// It returns an error if the storage operation fails.
func (s *NotificationService) AddSmtp(SMTPConfigs notification.CreateSMTPConfigRequest, userID uuid.UUID) error {
	s.logger.Log(logger.Info, "Adding SMTP configuration", "")
	config := notification.NewSMTPConfig(&SMTPConfigs, userID)
	return s.storage.AddSmtp(config)
}
