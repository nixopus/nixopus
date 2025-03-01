package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
)

func (s *NotificationService) UpdateSmtp(SMTPConfigs notification.UpdateSMTPConfigRequest) error {
	s.logger.Log(logger.Info, "Updating SMTP configuration", "")
	return s.storage.UpdateSmtp(&SMTPConfigs)
}
