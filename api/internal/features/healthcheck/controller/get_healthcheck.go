package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/healthcheck/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *HealthCheckController) GetHealthCheck(f fuego.ContextNoBody) (*types.HealthCheckResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		c.logger.Log(logger.Debug, "healthcheck: GetHealthCheck: authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == (uuid.UUID{}) {
		c.logger.Log(logger.Debug, "healthcheck: GetHealthCheck: organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	q := r.URL.Query()
	applicationID := q.Get("application_id")
	if applicationID == "" {
		c.logger.Log(logger.Debug, "healthcheck: GetHealthCheck: application_id required", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: types.ErrInvalidApplicationID.Error(), Err: types.ErrInvalidApplicationID}
	}

	ctxStr := fmt.Sprintf("org_id=%s user_id=%s application_id=%s", orgID, user.ID, applicationID)
	c.logger.Log(logger.Info, "healthcheck: GetHealthCheck", ctxStr)

	healthCheck, err := c.service.GetHealthCheck(applicationID, orgID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("healthcheck: GetHealthCheck: %v", err), ctxStr)
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	c.logger.Log(logger.Info, "healthcheck: GetHealthCheck ok", fmt.Sprintf("%s has_data=%t", ctxStr, healthCheck != nil))
	// Return success with null data if health check doesn't exist
	return &types.HealthCheckResponse{
		Status:  "success",
		Message: "Health check fetched successfully",
		Data:    healthCheck,
	}, nil
}
