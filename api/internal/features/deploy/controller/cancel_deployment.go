package controller

import (
	"io"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *DeployController) CancelDeployment(f fuego.ContextWithBody[types.CancelDeploymentRequest]) (*types.MessageResponse, error) {
	c.logger.Log(logger.Info, "starting deployment cancellation", "")

	data, err := f.Body()
	if err != nil {
		if err == io.EOF {
			c.logger.Log(logger.Error, "empty request body received", "deployment_id is required")
			return nil, fuego.BadRequestError{
				Detail: types.ErrMissingID.Error(),
				Err:    types.ErrMissingID,
			}
		}
		c.logger.Log(logger.Error, "failed to read request body", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	if err := c.validator.ValidateRequest(&data); err != nil {
		c.logger.Log(logger.Error, "request validation failed", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user authentication failed", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	deployment, err := c.storage.GetApplicationDeploymentById(data.DeploymentID.String())
	if err != nil {
		c.logger.Log(logger.Error, "deployment not found", err.Error())
		return nil, fuego.BadRequestError{
			Detail: types.ErrDeploymentNotRunning.Error(),
			Err:    types.ErrDeploymentNotRunning,
		}
	}

	if deployment.Status != nil {
		status := deployment.Status.Status
		if status != shared_types.Cloning && status != shared_types.Building && status != shared_types.Deploying && status != shared_types.Started {
			c.logger.Log(logger.Error, "deployment not in cancellable state", string(status))
			return nil, fuego.BadRequestError{
				Detail: types.ErrDeploymentNotCancellable.Error(),
				Err:    types.ErrDeploymentNotCancellable,
			}
		}
	}

	if err := c.taskService.CancelDeployment(data.DeploymentID.String()); err != nil {
		c.logger.Log(logger.Error, "failed to cancel deployment", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusBadRequest,
		}
	}

	c.logger.Log(logger.Info, "deployment cancelled successfully", data.DeploymentID.String())
	return &types.MessageResponse{
		Status:  "success",
		Message: "Deployment cancellation initiated",
	}, nil
}
