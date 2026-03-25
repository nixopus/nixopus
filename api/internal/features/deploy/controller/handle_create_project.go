package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// HandleCreateProject creates a new project (application) without triggering deployment.
// This allows users to save configuration as a draft and deploy later.
func (c *DeployController) HandleCreateProject(f fuego.ContextWithBody[types.CreateProjectRequest]) (*types.ApplicationResponse, error) {
	c.logger.Log(logger.Info, "creating project without deployment", "")

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

	c.logger.Log(logger.Info, "attempting to create project", "name: "+data.Name+", user_id: "+user.ID.String())

	application, err := c.service.CreateProject(&data, user.ID, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to create project", "name: "+data.Name+", error: "+err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "project created successfully", "name: "+data.Name)
	return &types.ApplicationResponse{
		Status:  "success",
		Message: "Project created successfully. You can deploy it when ready.",
		Data:    application,
	}, nil
}
