package service

import (
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetSmtp returns the SMTP configuration associated with the given ID.
//
// It logs an info message to the logger before calling the same method on the storage layer.
func (s *NotificationService) GetSmtp(ID string, organizationID string) (*shared_types.SMTPConfigs, error) {
	s.logger.Log(logger.Info, "Getting SMTP configuration", "")

	smtp, err := s.storage.GetSmtp(ID)
	if err == nil && smtp != nil {
		return smtp, nil
	}

	smtpConfigs, err := s.storage.GetOrganizationsSmtp(organizationID)
	if err != nil {
		return nil, err
	}

	if len(smtpConfigs) == 0 {
		return nil, fmt.Errorf("no SMTP configurations found for organization ID=%s", organizationID)
	}

	return &smtpConfigs[0], nil
}
