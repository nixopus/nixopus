package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *HealthCheckService) UpdateHealthCheck(organizationID uuid.UUID, req *types.UpdateHealthCheckRequest) (*shared_types.HealthCheck, error) {
	s.logger.Log(logger.Info, "updating health check", "application_id: "+req.ApplicationID)

	applicationID, err := uuid.Parse(req.ApplicationID)
	if err != nil {
		return nil, types.ErrInvalidApplicationID
	}

	healthCheck, err := s.storage.GetHealthCheckByApplicationID(applicationID, organizationID)
	if err != nil {
		return nil, types.ErrHealthCheckNotFound
	}

	// Update fields if provided
	if req.Endpoint != "" {
		healthCheck.Endpoint = req.Endpoint
	}

	if req.Method != "" {
		healthCheck.Method = req.Method
	}

	if len(req.ExpectedStatus) > 0 {
		healthCheck.ExpectedStatus = req.ExpectedStatus
	}

	if req.TimeoutSeconds > 0 {
		healthCheck.TimeoutSeconds = req.TimeoutSeconds
	}

	if req.IntervalSeconds > 0 {
		healthCheck.IntervalSeconds = req.IntervalSeconds
	}

	if req.FailureThreshold > 0 {
		healthCheck.FailureThreshold = req.FailureThreshold
	}

	if req.SuccessThreshold > 0 {
		healthCheck.SuccessThreshold = req.SuccessThreshold
	}

	if req.Headers != nil {
		healthCheck.Headers = req.Headers
	}

	if req.Body != "" {
		healthCheck.Body = req.Body
	}

	if req.RetentionDays > 0 {
		healthCheck.RetentionDays = req.RetentionDays
	}

	healthCheck.UpdatedAt = time.Now()

	if err := s.storage.UpdateHealthCheck(healthCheck); err != nil {
		s.logger.Log(logger.Error, "failed to update health check", err.Error())
		return nil, err
	}

	return healthCheck, nil
}
