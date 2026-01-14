package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *HealthCheckService) GetHealthCheckResults(applicationIDStr string, organizationID uuid.UUID, limit int, startTimeStr, endTimeStr string) ([]*shared_types.HealthCheckResult, error) {
	applicationID, err := uuid.Parse(applicationIDStr)
	if err != nil {
		return nil, types.ErrInvalidApplicationID
	}

	healthCheck, err := s.storage.GetHealthCheckByApplicationID(applicationID, organizationID)
	if err != nil {
		return nil, types.ErrHealthCheckNotFound
	}

	var startTime, endTime *time.Time

	if startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = &t
		}
	}

	if endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = &t
		}
	}

	if limit <= 0 {
		limit = 100
	}

	results, err := s.storage.GetHealthCheckResults(healthCheck.ID, limit, startTime, endTime)
	if err != nil {
		s.logger.Log(logger.Error, "failed to get health check results", err.Error())
		return nil, err
	}

	return results, nil
}
