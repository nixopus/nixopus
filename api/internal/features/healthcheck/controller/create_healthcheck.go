package controller

import (
	"fmt"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *HealthCheckController) CreateHealthCheck(f fuego.ContextWithBody[types.CreateHealthCheckRequest]) (*types.HealthCheckResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logger.Log(logger.Debug, "healthcheck: CreateHealthCheck: authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == (uuid.UUID{}) {
		c.logger.Log(logger.Debug, "healthcheck: CreateHealthCheck: organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("healthcheck: CreateHealthCheck body: %v", err), fmt.Sprintf("org_id=%s", orgID))
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	ctxStr := fmt.Sprintf("org_id=%s user_id=%s application_id=%s", orgID, user.ID, body.ApplicationID)
	c.logger.Log(logger.Info, "healthcheck: CreateHealthCheck", ctxStr)

	if err := c.validator.ValidateRequest(&body); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("healthcheck: CreateHealthCheck validation: %v", err), ctxStr)
		statusCode, mappedErr := mapHealthCheckError(err)
		return &types.HealthCheckResponse{
			Status: "error",
			Error:  mappedErr.Error(),
		}, fuego.HTTPError{Detail: mappedErr.Error(), Status: statusCode}
	}

	healthCheck, err := c.service.CreateHealthCheck(user.ID, orgID, &body)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("healthcheck: CreateHealthCheck: %v", err), ctxStr)
		statusCode, mappedErr := mapHealthCheckError(err)
		return &types.HealthCheckResponse{
			Status: "error",
			Error:  mappedErr.Error(),
		}, fuego.HTTPError{Detail: mappedErr.Error(), Status: statusCode}
	}

	c.logger.Log(logger.Info, "healthcheck: CreateHealthCheck ok", fmt.Sprintf("%s health_check_id=%s", ctxStr, healthCheck.ID))
	return &types.HealthCheckResponse{
		Status:  "success",
		Message: "Health check created successfully",
		Data:    healthCheck,
	}, nil
}
