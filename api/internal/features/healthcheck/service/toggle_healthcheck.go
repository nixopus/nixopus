package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (s *HealthCheckService) ToggleHealthCheck(organizationID uuid.UUID, req *types.ToggleHealthCheckRequest) (*shared_types.HealthCheck, error) {
	s.logger.Log(logger.Info, "healthcheck service: ToggleHealthCheck", fmt.Sprintf("application_id=%s org_id=%s enabled=%t", req.ApplicationID, organizationID, req.Enabled))

	applicationID, err := uuid.Parse(req.ApplicationID)
	if err != nil {
		return nil, types.ErrInvalidApplicationID
	}

	if err := s.storage.ToggleHealthCheck(applicationID, organizationID, req.Enabled); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("healthcheck service: ToggleHealthCheck: %v", err), fmt.Sprintf("application_id=%s org_id=%s", req.ApplicationID, organizationID))
		return nil, err
	}

	healthCheck, err := s.storage.GetHealthCheckByApplicationID(applicationID, organizationID)
	if err != nil {
		return nil, types.ErrHealthCheckNotFound
	}

	return healthCheck, nil
}
