package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// GetApplicationServers returns all servers assigned to an application.
func (c *DeployController) GetApplicationServers(f fuego.ContextNoBody) (*types.ApplicationServersResponse, error) {
	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "organization not found",
		}
	}

	appIDStr := f.QueryParam("id")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		c.logger.Log(logger.Error, "invalid application id", err.Error())
		return nil, fuego.BadRequestError{
			Detail: "invalid application id",
			Err:    err,
		}
	}

	if _, err := c.storage.GetApplicationById(appID.String(), organizationID); err != nil {
		c.logger.Log(logger.Error, "application not found or not authorized", err.Error())
		return nil, fuego.NotFoundError{
			Detail: "application not found",
		}
	}

	servers, err := c.storage.GetApplicationServers(appID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ApplicationServersResponse{
		Status:  "success",
		Message: "Application servers retrieved successfully",
		Data:    servers,
	}, nil
}

// SetApplicationServers replaces the server assignment for an application.
func (c *DeployController) SetApplicationServers(f fuego.ContextWithBody[types.SetApplicationServersRequest]) (*types.ApplicationServersResponse, error) {
	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "authentication required",
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "")
		return nil, fuego.UnauthorizedError{
			Detail: "organization not found",
		}
	}

	data, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, "failed to read request body", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	if len(data.ServerIDs) == 0 {
		c.logger.Log(logger.Error, "server_ids is required", "")
		return nil, fuego.BadRequestError{
			Detail: types.ErrAtLeastOneServerRequired.Error(),
			Err:    types.ErrAtLeastOneServerRequired,
		}
	}

	if data.PrimaryServerID != nil {
		found := false
		for _, sid := range data.ServerIDs {
			if sid == *data.PrimaryServerID {
				found = true
				break
			}
		}
		if !found {
			c.logger.Log(logger.Error, "primary_server_id must be in server_ids", "")
			return nil, fuego.BadRequestError{
				Detail: "primary_server_id must be one of the provided server_ids",
			}
		}
	}

	if _, err := c.storage.GetApplicationById(data.ApplicationID.String(), organizationID); err != nil {
		c.logger.Log(logger.Error, "application not found or not authorized", err.Error())
		return nil, fuego.NotFoundError{
			Detail: "application not found",
		}
	}

	if err := c.storage.SetApplicationServers(data.ApplicationID, data.ServerIDs, data.PrimaryServerID, data.RoutingStrategy); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	servers, err := c.storage.GetApplicationServers(data.ApplicationID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to reload application servers", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ApplicationServersResponse{
		Status:  "success",
		Message: "Application servers updated successfully",
		Data:    servers,
	}, nil
}
