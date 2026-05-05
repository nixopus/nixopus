package service

import (
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (s *HealthCheckService) GetDueHealthChecks() ([]*shared_types.HealthCheck, error) {
	checks, err := s.storage.GetDueHealthChecks()
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("healthcheck service: GetDueHealthChecks: %v", err), "")
		return nil, err
	}
	return checks, nil
}
