package controller

import (
	"fmt"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *HealthCheckController) DeleteHealthCheck(f fuego.ContextNoBody) (*types.HealthCheckMessageResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logger.Log(logger.Debug, "healthcheck: DeleteHealthCheck: authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == (uuid.UUID{}) {
		c.logger.Log(logger.Debug, "healthcheck: DeleteHealthCheck: organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	q := r.URL.Query()
	applicationID := q.Get("application_id")
	if applicationID == "" {
		c.logger.Log(logger.Debug, "healthcheck: DeleteHealthCheck: application_id required", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: types.ErrInvalidApplicationID.Error(), Err: types.ErrInvalidApplicationID}
	}

	ctxStr := fmt.Sprintf("org_id=%s user_id=%s application_id=%s", orgID, user.ID, applicationID)
	c.logger.Log(logger.Info, "healthcheck: DeleteHealthCheck", ctxStr)

	if err := c.service.DeleteHealthCheck(applicationID, orgID); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("healthcheck: DeleteHealthCheck: %v", err), ctxStr)
		statusCode, mappedErr := mapHealthCheckError(err)
		return &types.HealthCheckMessageResponse{
			Status: "error",
			Error:  mappedErr.Error(),
		}, fuego.HTTPError{Detail: mappedErr.Error(), Status: statusCode}
	}

	c.logger.Log(logger.Info, "healthcheck: DeleteHealthCheck ok", ctxStr)
	return &types.HealthCheckMessageResponse{
		Status:  "success",
		Message: "Health check deleted successfully",
	}, nil
}
