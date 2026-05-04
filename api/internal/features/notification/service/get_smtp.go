package service

import (
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// GetSmtp returns the SMTP configuration associated with the given ID.
//
// It logs an info message to the logger before calling the same method on the storage layer.
func (s *NotificationService) GetSmtp(ID string, organizationID string) (*shared_types.SMTPConfigs, error) {
	data := fmt.Sprintf("query_user_id=%s org_id=%s", ID, organizationID)
	s.logger.Log(logger.Info, "notification service: GetSmtp", data)

	smtp, err := s.storage.GetSmtp(ID)
	if err != nil {
		return nil, err
	}
	if smtp != nil {
		return smtp, nil
	}

	smtpConfigs, err := s.storage.GetOrganizationsSmtp(organizationID)
	if err != nil {
		s.logger.Log(logger.Warning, fmt.Sprintf("notification service: GetSmtp org lookup failed (migration pending?): %v", err), data)
		return nil, notification.ErrSMTPConfigNotFound
	}

	if len(smtpConfigs) == 0 {
		return nil, notification.ErrSMTPConfigNotFound
	}

	return &smtpConfigs[0], nil
}
