package controller

import (
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

	if err := c.service.DeleteHealthCheck(applicationID, orgID); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		statusCode, mappedErr := mapHealthCheckError(err)
		return &types.HealthCheckMessageResponse{
			Status: "error",
			Error:  mappedErr.Error(),
		}, fuego.HTTPError{Detail: mappedErr.Error(), Status: statusCode}
	}

	return &types.HealthCheckMessageResponse{
		Status:  "success",
		Message: "Health check deleted successfully",
	}, nil
}
