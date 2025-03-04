package service

import "github.com/raghavyuva/nixopus-api/internal/features/logger"

// DeleteSmtp deletes a SMTP configuration.
//
// It takes an ID as a parameter and calls the same method on the storage layer.
//
// It logs an info message to the logger before calling the storage layer.
func (s *NotificationService) DeleteSmtp(ID string) error {
	s.logger.Log(logger.Info, "Deleting SMTP configuration", "")
	return s.storage.DeleteSmtp(ID)
}
