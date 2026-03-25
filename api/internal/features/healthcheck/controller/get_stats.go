package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	healthcheck_service "github.com/nixopus/nixopus/api/internal/features/healthcheck/service"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *HealthCheckController) GetHealthCheckStats(f fuego.ContextNoBody) (*types.HealthCheckStatsResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == (uuid.UUID{}) {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	q := r.URL.Query()
	applicationID := q.Get("application_id")
	if applicationID == "" {
		return nil, fuego.BadRequestError{Detail: types.ErrInvalidApplicationID.Error(), Err: types.ErrInvalidApplicationID}
	}

	period := q.Get("period")
	if period == "" {
		period = "24h"
	}

	stats, err := c.service.GetHealthCheckStats(applicationID, orgID, period)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
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
