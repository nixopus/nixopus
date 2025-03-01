package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *NotificationService) GetSmtp(ID string) (*shared_types.SMTPConfigs, error) {
	s.logger.Log(logger.Info, "Getting SMTP configuration", "")
	return s.storage.GetSmtp(ID)
}
