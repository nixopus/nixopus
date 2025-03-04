package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetSmtp returns the SMTP configuration associated with the given ID.
//
// It logs an info message to the logger before calling the same method on the storage layer.
func (s *NotificationService) GetSmtp(ID string) (*shared_types.SMTPConfigs, error) {
	s.logger.Log(logger.Info, "Getting SMTP configuration", "")
	return s.storage.GetSmtp(ID)
}
