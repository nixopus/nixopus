package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type HealthCheckStatsResponse struct {
	ApplicationID    string  `json:"application_id"`
	UptimePercentage float64 `json:"uptime_percentage"`
	AvgResponseTime  int     `json:"avg_response_time_ms"`
	TotalChecks      int     `json:"total_checks"`
	SuccessfulChecks int     `json:"successful_checks"`
	FailedChecks     int     `json:"failed_checks"`
	Period           string  `json:"period"`
	LastStatus       string  `json:"last_status"`
}

func (s *HealthCheckService) GetHealthCheckStats(applicationIDStr string, organizationID uuid.UUID, period string) (*HealthCheckStatsResponse, error) {
	applicationID, err := uuid.Parse(applicationIDStr)
	if err != nil {
		return nil, types.ErrInvalidApplicationID
	}

	healthCheck, err := s.storage.GetHealthCheckByApplicationID(applicationID, organizationID)
	if err != nil {
		return nil, types.ErrHealthCheckNotFound
	}

	// Parse period (default to 24h)
	var duration time.Duration
	switch period {
	case "1h", "1H":
		duration = 1 * time.Hour
	case "24h", "24H", "1d", "1D":
		duration = 24 * time.Hour
	case "7d", "7D":
		duration = 7 * 24 * time.Hour
	case "30d", "30D":
		duration = 30 * 24 * time.Hour
	default:
		duration = 24 * time.Hour
		period = "24h"
	}

	endTime := time.Now()
	startTime := endTime.Add(-duration)

	stats, err := s.storage.GetHealthCheckStats(healthCheck.ID, startTime, endTime)
	if err != nil {
		s.logger.Log(logger.Error, "failed to get health check stats", err.Error())
		return nil, err
	}

	// Get last status
	results, err := s.storage.GetHealthCheckResults(healthCheck.ID, 1, nil, nil)
	lastStatus := string(shared_types.HealthCheckStatusUnknown)
	if err == nil && len(results) > 0 {
		lastStatus = results[0].Status
	}

	return &HealthCheckStatsResponse{
		ApplicationID:    applicationIDStr,
		UptimePercentage: stats.UptimePercentage,
		AvgResponseTime:  stats.AvgResponseTime,
		TotalChecks:      stats.TotalChecks,
		SuccessfulChecks: stats.SuccessfulChecks,
		FailedChecks:     stats.FailedChecks,
		Period:           period,
		LastStatus:       lastStatus,
	}, nil
}
