package service

import (
	"context"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *DeployService) GetDeploymentLogs(ctx context.Context, deploymentID string, page, pageSize int, level string, startTime, endTime time.Time, searchTerm string) ([]shared_types.ApplicationLogs, int64, error) {
	logs, totalCount, err := s.storage.GetDeploymentLogs(deploymentID, page, pageSize, level, startTime, endTime, searchTerm)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get deployment logs", err.Error())
		return nil, 0, err
	}

	return logs, int64(totalCount), nil
}
