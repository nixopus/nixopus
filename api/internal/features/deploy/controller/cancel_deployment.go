package controller

import (
	"fmt"
	"io"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *DeployController) CancelDeployment(f fuego.ContextWithBody[types.CancelDeploymentRequest]) (*types.MessageResponse, error) {
	c.logger.Log(logger.Info, "deploy: starting deployment cancellation", "")

	data, err := f.Body()
	if err != nil {
		if err == io.EOF {
			c.logger.Log(logger.Error, "deploy: empty request body received", "deployment_id is required")
			return nil, fuego.BadRequestError{
				Detail: types.ErrMissingID.Error(),
				Err:    types.ErrMissingID,
			}
		}
		c.logger.Log(logger.Error, "deploy: failed to read request body", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	if err := c.validator.ValidateRequest(&data); err != nil {
		c.logger.Log(logger.Error, "deploy: request validation failed", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "deploy: user authentication failed", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "deploy: organization not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "organization not found",
		}
	}

	deployment, err := c.storage.GetApplicationDeploymentById(data.DeploymentID.String())
	if err != nil {
		c.logger.Log(logger.Error, "deploy: deployment not found", err.Error())
		return nil, fuego.BadRequestError{
			Detail: types.ErrDeploymentNotRunning.Error(),
			Err:    types.ErrDeploymentNotRunning,
		}
	}

	if deployment.Application == nil || deployment.Application.OrganizationID != organizationID {
		c.logger.Log(logger.Error, "deploy: deployment not owned by caller", fmt.Sprintf("deployment=%s org=%s", deployment.ID, organizationID))
		return nil, fuego.ForbiddenError{
			Detail: "you do not have permission to cancel this deployment",
			Err:    types.ErrPermissionDenied,
		}
	}

	if deployment.Status == nil {
		c.logger.Log(logger.Error, "deploy: deployment status missing", data.DeploymentID.String())
		return nil, fuego.BadRequestError{
			Detail: types.ErrDeploymentNotCancellable.Error(),
			Err:    types.ErrDeploymentNotCancellable,
		}
	}

	status := deployment.Status.Status
	if status != shared_types.Cloning && status != shared_types.Building && status != shared_types.Deploying && status != shared_types.Started {
		c.logger.Log(logger.Error, "deploy: deployment not in cancellable state", string(status))
		return nil, fuego.BadRequestError{
			Detail: types.ErrDeploymentNotCancellable.Error(),
			Err:    types.ErrDeploymentNotCancellable,
		}
	}

	if err := c.taskService.CancelDeployment(data.DeploymentID.String()); err != nil {
		c.logger.Log(logger.Error, "deploy: failed to cancel deployment", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusBadRequest,
		}
	}

	c.logger.Log(logger.Info, "deploy: deployment cancelled successfully", data.DeploymentID.String())
	return &types.MessageResponse{
		Status:  "success",
		Message: "Deployment cancellation initiated",
	}, nil
}
