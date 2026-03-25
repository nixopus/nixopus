package controller

import (
	"io"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *DeployController) UpdateApplication(f fuego.ContextWithBody[types.UpdateDeploymentRequest]) (*types.ApplicationResponse, error) {
	c.logger.Log(logger.Info, "starting application update process", "")

	data, err := f.Body()
	if err != nil {
		if err == io.EOF {
			c.logger.Log(logger.Error, "empty request body received", "id is required for update")
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

	c.logger.Log(logger.Info, "request body parsed successfully", "id: "+data.ID.String())

	if err := c.validator.ValidateRequest(&data); err != nil {
		c.logger.Log(logger.Error, "request validation failed", "id: "+data.ID.String()+", error: "+err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user authentication failed", "id: "+data.ID.String())
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "id: "+data.ID.String())
		return nil, fuego.UnauthorizedError{
			Detail: "organization not found",
		}
	}

	c.logger.Log(logger.Info, "attempting to update application", "id: "+data.ID.String()+", user_id: "+user.ID.String())

	// Sync compose-specific domains if provided, otherwise sync plain domains
	if len(data.ComposeDomains) > 0 {
		if err := c.syncComposeApplicationDomains(data.ID, organizationID, data.ComposeDomains); err != nil {
			c.logger.Log(logger.Error, "failed to sync compose application domains", err.Error())
			return nil, fuego.BadRequestError{
				Detail: err.Error(),
				Err:    err,
			}
		}
	} else if data.Domains != nil {
		if err := c.syncApplicationDomains(data.ID, organizationID, data.Domains); err != nil {
			c.logger.Log(logger.Error, "failed to sync application domains", err.Error())
			return nil, fuego.BadRequestError{
				Detail: err.Error(),
				Err:    err,
			}
		}
	}

	application, err := c.taskService.UpdateDeployment(&data, user.ID, organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to create deployment", "name: "+data.Name+", error: "+err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	// Reload application with domains for response
	application, err = c.service.GetApplicationById(data.ID.String(), organizationID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to reload application", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "application updated successfully", "id: "+data.ID.String())
	return &types.ApplicationResponse{
		Status:  "success",
		Message: "Application updated successfully",
		Data:    application,
	}, nil
}
