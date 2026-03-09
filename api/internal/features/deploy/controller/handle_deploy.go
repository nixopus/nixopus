package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *DeployController) HandleDeploy(f fuego.ContextWithBody[types.CreateDeploymentRequest]) (*types.ApplicationResponse, error) {
	c.logger.Log(logger.Info, "starting deployment process", "")

	data, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, "failed to read request body", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	c.logger.Log(logger.Info, "request body parsed successfully", "name: "+data.Name)

	if err := c.validator.ValidateRequest(&data); err != nil {
		c.logger.Log(logger.Error, "request validation failed", "name: "+data.Name+", error: "+err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user authentication failed", "name: "+data.Name)
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "name: "+data.Name)
		return nil, fuego.UnauthorizedError{
			Detail: "organization not found",
		}
	}

	c.logger.Log(logger.Info, "attempting to create deployment", "name: "+data.Name+", user_id: "+user.ID.String())

	application, err := c.taskService.CreateDeploymentTask(&data, user.ID, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to create deployment", "name: "+data.Name+", error: "+err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	// application, err := c.service.CreateDeployment(&data, user.ID, organizationID)
	// if err != nil {
	// 	c.logger.Log(logger.Error, "failed to create deployment", "name: "+data.Name+", error: "+err.Error())
	// 	return nil, fuego.HTTPError{
	// 		Err:    err,
	// 		Status: http.StatusInternalServerError,
	// 	}
	// }

	c.logger.Log(logger.Info, "deployment created successfully", "name: "+data.Name)
	return &types.ApplicationResponse{
		Status:  "success",
		Message: "Deployment created successfully",
		Data:    application,
	}, nil
}
