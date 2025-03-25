package service

import (
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetSmtp returns the SMTP configuration associated with the given ID.
//
// It logs an info message to the logger before calling the same method on the storage layer.
func (s *NotificationService) GetSmtp(ID string,organizationID string) (*shared_types.SMTPConfigs, error) {
	s.logger.Log(logger.Info, "Getting SMTP configuration", "")
	smtp,err:= s.storage.GetOrganizationsSmtp(organizationID)
	if err!=nil{
		return nil, err
	}

	if len(smtp) == 0 {
		return nil, fmt.Errorf("no SMTP configurations with ID=%s found", organizationID)
	}

	return &smtp[0], nil
}
