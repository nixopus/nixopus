package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *HealthCheckService) ToggleHealthCheck(organizationID uuid.UUID, req *types.ToggleHealthCheckRequest) (*shared_types.HealthCheck, error) {
	s.logger.Log(logger.Info, "toggling health check", fmt.Sprintf("application_id: %s, enabled: %t", req.ApplicationID, req.Enabled))

	applicationID, err := uuid.Parse(req.ApplicationID)
	if err != nil {
		return nil, types.ErrInvalidApplicationID
	}

	if err := s.storage.ToggleHealthCheck(applicationID, organizationID, req.Enabled); err != nil {
		s.logger.Log(logger.Error, "failed to toggle health check", err.Error())
		return nil, err
	}

	healthCheck, err := s.storage.GetHealthCheckByApplicationID(applicationID, organizationID)
	if err != nil {
		return nil, types.ErrHealthCheckNotFound
	}

	return healthCheck, nil
}
