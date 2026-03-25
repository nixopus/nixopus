package service

import (
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (s *HealthCheckService) GetHealthCheck(applicationIDStr string, organizationID uuid.UUID) (*shared_types.HealthCheck, error) {
	applicationID, err := uuid.Parse(applicationIDStr)
	if err != nil {
		return nil, types.ErrInvalidApplicationID
	}

	healthCheck, err := s.storage.GetHealthCheckByApplicationID(applicationID, organizationID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Return nil, nil when health check is not found (not an error)
			return nil, nil
		}
		s.logger.Log(logger.Error, "failed to get health check", err.Error())
		return nil, types.ErrHealthCheckNotFound
	}

	return healthCheck, nil
}
