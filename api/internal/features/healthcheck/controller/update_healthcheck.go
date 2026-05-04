package controller

import (
	"fmt"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *HealthCheckController) UpdateHealthCheck(f fuego.ContextWithBody[types.UpdateHealthCheckRequest]) (*types.HealthCheckResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logger.Log(logger.Debug, "healthcheck: UpdateHealthCheck: authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == (uuid.UUID{}) {
		c.logger.Log(logger.Debug, "healthcheck: UpdateHealthCheck: organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("healthcheck: UpdateHealthCheck body: %v", err), fmt.Sprintf("org_id=%s", orgID))
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	ctxStr := fmt.Sprintf("org_id=%s user_id=%s application_id=%s", orgID, user.ID, body.ApplicationID)
	c.logger.Log(logger.Info, "healthcheck: UpdateHealthCheck", ctxStr)

	if err := c.validator.ValidateRequest(&body); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("healthcheck: UpdateHealthCheck validation: %v", err), ctxStr)
		statusCode, mappedErr := mapHealthCheckError(err)
		return &types.HealthCheckResponse{
			Status: "error",
			Error:  mappedErr.Error(),
		}, fuego.HTTPError{Detail: mappedErr.Error(), Status: statusCode}
	}

	healthCheck, err := c.service.UpdateHealthCheck(orgID, &body)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("healthcheck: UpdateHealthCheck: %v", err), ctxStr)
		statusCode, mappedErr := mapHealthCheckError(err)
		return &types.HealthCheckResponse{
			Status: "error",
			Error:  mappedErr.Error(),
		}, fuego.HTTPError{Detail: mappedErr.Error(), Status: statusCode}
	}

	c.logger.Log(logger.Info, "healthcheck: UpdateHealthCheck ok", fmt.Sprintf("%s health_check_id=%s", ctxStr, healthCheck.ID))
	return &types.HealthCheckResponse{
		Status:  "success",
		Message: "Health check updated successfully",
		Data:    healthCheck,
	}, nil
}
