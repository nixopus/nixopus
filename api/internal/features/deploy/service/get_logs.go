package service

import (
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *DeployService) GetLogs(applicationID string, page, pageSize int, level string, startTime, endTime time.Time, searchTerm string) ([]shared_types.ApplicationLogs, int, error) {
	logs, totalCount, err := s.storage.GetLogs(applicationID, page, pageSize, level, startTime, endTime, searchTerm)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to get logs", err.Error())
		return nil, 0, err
	}

	return logs, totalCount, nil
}
