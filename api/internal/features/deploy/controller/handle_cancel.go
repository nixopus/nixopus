package controller

import (
	"io"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// HandleCancel cancels an ongoing deployment.
// It updates the deployment status to cancelled if the deployment is in progress.
func (c *DeployController) HandleCancel(f fuego.ContextWithBody[types.CancelDeploymentRequest]) (*types.MessageResponse, error) {
	c.logger.Log(logger.Info, "starting deployment cancellation process", "")

	data, err := f.Body()
	if err != nil {
		if err == io.EOF {
			c.logger.Log(logger.Error, "empty request body received", "id is required for cancel")
			return nil, fuego.HTTPError{
				Err:    types.ErrMissingID,
				Status: http.StatusBadRequest,
			}
		}
		c.logger.Log(logger.Error, "failed to read request body", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	c.logger.Log(logger.Info, "request body parsed successfully", "id: "+data.ID.String())

	if err := c.validator.ValidateRequest(&data); err != nil {
		c.logger.Log(logger.Error, "request validation failed", "id: "+data.ID.String()+", error: "+err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user authentication failed", "id: "+data.ID.String())
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	c.logger.Log(logger.Info, "attempting to cancel deployment", "id: "+data.ID.String()+", user_id: "+user.ID.String())

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "id: "+data.ID.String())
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	err = c.service.CancelDeployment(data.ID, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to cancel deployment", "id: "+data.ID.String()+", error: "+err.Error())
		status := http.StatusInternalServerError
		if err == types.ErrDeploymentNotInProgress || err == types.ErrDeploymentAlreadyCancelled {
			status = http.StatusBadRequest
		}
		return nil, fuego.HTTPError{
			Err:    err,
			Status: status,
		}
	}

	c.logger.Log(logger.Info, "deployment cancelled successfully", "id: "+data.ID.String())
	return &types.MessageResponse{
		Status:  "success",
		Message: "Deployment cancelled successfully",
	}, nil
}
