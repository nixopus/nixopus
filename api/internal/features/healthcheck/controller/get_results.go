package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *HealthCheckController) GetHealthCheckResults(f fuego.ContextNoBody) (*types.HealthCheckResultsResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logger.Log(logger.Debug, "healthcheck: GetHealthCheckResults: authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == (uuid.UUID{}) {
		c.logger.Log(logger.Debug, "healthcheck: GetHealthCheckResults: organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	q := r.URL.Query()
	applicationID := q.Get("application_id")
	if applicationID == "" {
		c.logger.Log(logger.Debug, "healthcheck: GetHealthCheckResults: application_id required", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: types.ErrInvalidApplicationID.Error(), Err: types.ErrInvalidApplicationID}
	}

	limit := 100
	if limitStr := q.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	startTime := q.Get("start_time")
	endTime := q.Get("end_time")

	ctxStr := fmt.Sprintf("org_id=%s user_id=%s application_id=%s limit=%d", orgID, user.ID, applicationID, limit)
	c.logger.Log(logger.Info, "healthcheck: GetHealthCheckResults", ctxStr)

	results, err := c.service.GetHealthCheckResults(applicationID, orgID, limit, startTime, endTime)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("healthcheck: GetHealthCheckResults: %v", err), ctxStr)
		statusCode, mappedErr := mapHealthCheckError(err)
		if statusCode == http.StatusNotFound {
			// Plain HTTPError with Status set — reliable with Fuego's castHTTPError / StatusCode().
			return nil, fuego.HTTPError{
				Status: http.StatusNotFound,
				Title:  "Not Found",
				Detail: mappedErr.Error(),
				Err:    mappedErr,
			}
		}
		return nil, fuego.HTTPError{Detail: mappedErr.Error(), Status: statusCode}
	}

	c.logger.Log(logger.Info, "healthcheck: GetHealthCheckResults ok", fmt.Sprintf("%s count=%d", ctxStr, len(results)))
	return &types.HealthCheckResultsResponse{
		Status:  "success",
		Message: "Health check results fetched successfully",
		Data:    results,
	}, nil
}
