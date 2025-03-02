package service

import "github.com/raghavyuva/nixopus-api/internal/features/logger"

func (s *NotificationService) DeleteSmtp(ID string) error {
	s.logger.Log(logger.Info, "Deleting SMTP configuration", "")
	return s.storage.DeleteSmtp(ID)
}
