package controller

import (
	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/trail/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// ProvisionTrail handles POST /api/v1/trail/provision
//
// This endpoint initiates a new trail provisioning request for the authenticated user.
// The request is validated and a task is enqueued for async processing.
//
// Request Body:
//   - image (optional): the container image to use
//
// Returns:
//   - 202 Accepted: provisioning started successfully
//   - 400 Bad Request: invalid request body or image not allowed
//   - 401 Unauthorized: authentication required
//   - 403 Forbidden: organization context required
//   - 409 Conflict: active provision already exists
//   - 503 Service Unavailable: system at capacity
func (c *TrailController) ProvisionTrail(f fuego.ContextWithBody[types.ProvisionRequest]) (*types.ProvisionTrailResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := r.Header.Get("X-Organization-Id")
	if orgID == "" {
		return nil, fuego.ForbiddenError{Detail: types.ErrOrganizationRequired.Error(), Err: types.ErrOrganizationRequired}
	}

	if _, err := uuid.Parse(orgID); err != nil {
		return nil, fuego.BadRequestError{Detail: types.ErrInvalidOrganizationID.Error(), Err: types.ErrInvalidOrganizationID}
	}

	body, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), user.ID.String())
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if err := c.validator.ValidateRequest(&body); err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	result, err := c.service.ProvisionTrail(user.ID.String(), orgID, body)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), user.ID.String())
		status := mapErrorToStatus(err)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: status,
		}
	}

	return &types.ProvisionTrailResponse{
		Status:  "success",
		Message: "Trail provisioning started",
		Data:    result,
	}, nil
}
