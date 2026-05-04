package controller

import (
	"fmt"
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
		c.logger.Log(logger.Debug, "healthcheck: GetHealthCheckStats: authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == (uuid.UUID{}) {
		c.logger.Log(logger.Debug, "healthcheck: GetHealthCheckStats: organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	q := r.URL.Query()
	applicationID := q.Get("application_id")
	if applicationID == "" {
		c.logger.Log(logger.Debug, "healthcheck: GetHealthCheckStats: application_id required", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: types.ErrInvalidApplicationID.Error(), Err: types.ErrInvalidApplicationID}
	}

	period := q.Get("period")
	if period == "" {
		period = "24h"
	}

	ctxStr := fmt.Sprintf("org_id=%s user_id=%s application_id=%s period=%s", orgID, user.ID, applicationID, period)
	c.logger.Log(logger.Info, "healthcheck: GetHealthCheckStats", ctxStr)

	stats, err := c.service.GetHealthCheckStats(applicationID, orgID, period)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("healthcheck: GetHealthCheckStats: %v", err), ctxStr)
		statusCode, mappedErr := mapHealthCheckError(err)
		if statusCode == http.StatusNotFound {
			return nil, fuego.HTTPError{
				Status: http.StatusNotFound,
				Title:  "Not Found",
				Detail: mappedErr.Error(),
				Err:    mappedErr,
			}
		}
		return nil, fuego.HTTPError{Detail: mappedErr.Error(), Status: statusCode}
	}

	c.logger.Log(logger.Info, "healthcheck: GetHealthCheckStats ok", ctxStr)
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
