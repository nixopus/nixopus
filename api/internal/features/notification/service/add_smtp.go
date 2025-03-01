package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
)

func (s *NotificationService) AddSmtp(SMTPConfigs notification.CreateSMTPConfigRequest, userID uuid.UUID) error {
	s.logger.Log(logger.Info, "Adding SMTP configuration", "")
	config := notification.NewSMTPConfig(&SMTPConfigs, userID)
	return s.storage.AddSmtp(config)
}
