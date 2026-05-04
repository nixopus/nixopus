package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (s *HealthCheckService) DeleteHealthCheck(applicationIDStr string, organizationID uuid.UUID) error {
	s.logger.Log(logger.Info, "healthcheck service: DeleteHealthCheck", fmt.Sprintf("application_id=%s org_id=%s", applicationIDStr, organizationID))

	applicationID, err := uuid.Parse(applicationIDStr)
	if err != nil {
		return types.ErrInvalidApplicationID
	}

	if err := s.storage.DeleteHealthCheck(applicationID, organizationID); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("healthcheck service: DeleteHealthCheck: %v", err), fmt.Sprintf("application_id=%s org_id=%s", applicationIDStr, organizationID))
		return err
	}

	return nil
}
