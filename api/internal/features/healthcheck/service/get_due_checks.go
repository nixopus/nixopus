package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *HealthCheckService) GetDueHealthChecks() ([]*shared_types.HealthCheck, error) {
	checks, err := s.storage.GetDueHealthChecks()
	if err != nil {
		s.logger.Log(logger.Error, "failed to get due health checks", err.Error())
		return nil, err
	}
	return checks, nil
}
