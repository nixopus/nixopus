package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/healthcheck/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *HealthCheckController) UpdateHealthCheck(f fuego.ContextWithBody[types.UpdateHealthCheckRequest]) (*shared_types.Response, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{Status: http.StatusUnauthorized}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == (uuid.UUID{}) {
		return nil, fuego.HTTPError{Status: http.StatusBadRequest}
	}

	body, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
	}

	if err := c.validator.ValidateRequest(&body); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		statusCode, mappedErr := mapHealthCheckError(err)
		return &shared_types.Response{
			Status: "error",
			Error:  mappedErr.Error(),
		}, fuego.HTTPError{Status: statusCode}
	}

	healthCheck, err := c.service.UpdateHealthCheck(orgID, &body)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		statusCode, mappedErr := mapHealthCheckError(err)
		return &shared_types.Response{
			Status: "error",
			Error:  mappedErr.Error(),
		}, fuego.HTTPError{Status: statusCode}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Health check updated successfully",
		Data:    healthCheck,
	}, nil
}
