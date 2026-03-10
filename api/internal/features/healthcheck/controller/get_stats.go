package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	healthcheck_service "github.com/raghavyuva/nixopus-api/internal/features/healthcheck/service"
	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *HealthCheckController) GetHealthCheckStats(f fuego.ContextNoBody) (*types.HealthCheckStatsResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{Status: http.StatusUnauthorized}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == (uuid.UUID{}) {
		return nil, fuego.HTTPError{Status: http.StatusBadRequest}
	}

	q := r.URL.Query()
	applicationID := q.Get("application_id")
	if applicationID == "" {
		return nil, fuego.HTTPError{Status: http.StatusBadRequest, Err: types.ErrInvalidApplicationID}
	}

	period := q.Get("period")
	if period == "" {
		period = "24h"
	}

	stats, err := c.service.GetHealthCheckStats(applicationID, orgID, period)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Status: http.StatusInternalServerError}
	}

	data := mapStatsResponse(stats)

	return &types.HealthCheckStatsResponse{
		Status:  "success",
		Message: "Health check stats fetched successfully",
		Data:    data,
	}, nil
}

func mapStatsResponse(stats *healthcheck_service.HealthCheckStatsResponse) *types.HealthCheckStatsData {
	if stats == nil {
		return nil
	}

	return &types.HealthCheckStatsData{
		ApplicationID:    stats.ApplicationID,
		UptimePercentage: stats.UptimePercentage,
		AvgResponseTime:  stats.AvgResponseTime,
		TotalChecks:      stats.TotalChecks,
		SuccessfulChecks: stats.SuccessfulChecks,
		FailedChecks:     stats.FailedChecks,
		Period:           stats.Period,
		LastStatus:       stats.LastStatus,
	}
}
