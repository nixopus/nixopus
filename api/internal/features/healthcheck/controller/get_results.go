package controller

import (
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

	limit := 100
	if limitStr := q.Get("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	startTime := q.Get("start_time")
	endTime := q.Get("end_time")

	results, err := c.service.GetHealthCheckResults(applicationID, orgID, limit, startTime, endTime)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
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

	return &types.HealthCheckResultsResponse{
		Status:  "success",
		Message: "Health check results fetched successfully",
		Data:    results,
	}, nil
}
