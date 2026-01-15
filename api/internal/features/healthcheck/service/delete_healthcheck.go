package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (s *HealthCheckService) DeleteHealthCheck(applicationIDStr string, organizationID uuid.UUID) error {
	s.logger.Log(logger.Info, "deleting health check", "application_id: "+applicationIDStr)

	applicationID, err := uuid.Parse(applicationIDStr)
	if err != nil {
		return types.ErrInvalidApplicationID
	}

	if err := s.storage.DeleteHealthCheck(applicationID, organizationID); err != nil {
		s.logger.Log(logger.Error, "failed to delete health check", err.Error())
		return err
	}

	return nil
}
