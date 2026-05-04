package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/caddy"
	"github.com/nixopus/nixopus/api/internal/features/deploy/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// GetApplicationServers returns all servers assigned to an application.
func (c *DeployController) GetApplicationServers(f fuego.ContextNoBody) (*types.ApplicationServersResponse, error) {
	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "deploy: user not found", "")
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

	appIDStr := f.QueryParam("id")
	appID, err := uuid.Parse(appIDStr)
	if err != nil {
		c.logger.Log(logger.Error, "deploy: invalid application id", err.Error())
		return nil, fuego.BadRequestError{
			Detail: "invalid application id",
			Err:    err,
		}
	}

	if _, err := c.storage.GetApplicationById(appID.String(), organizationID); err != nil {
		c.logger.Log(logger.Error, "deploy: application not found or not authorized", err.Error())
		return nil, fuego.NotFoundError{
			Detail: "application not found",
		}
	}

	logData := deployRequestData(f.Request(), user)
	if logData == "" {
		logData = fmt.Sprintf("application_id=%s", appID.String())
	} else {
		logData = fmt.Sprintf("%s application_id=%s", logData, appID.String())
	}

	servers, err := c.storage.GetApplicationServers(appID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("deploy: GetApplicationServers: %v", err), logData)
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
		c.logger.Log(logger.Error, "deploy: user not found", "")
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

	data, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, "deploy: failed to read request body", err.Error())
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	if len(data.ServerIDs) == 0 {
		c.logger.Log(logger.Error, "deploy: server_ids is required", "")
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
			c.logger.Log(logger.Error, "deploy: primary_server_id must be in server_ids", "")
			return nil, fuego.BadRequestError{
				Detail: "primary_server_id must be one of the provided server_ids",
			}
		}
	}

	if _, err := c.storage.GetApplicationById(data.ApplicationID.String(), organizationID); err != nil {
		c.logger.Log(logger.Error, "deploy: application not found or not authorized", err.Error())
		return nil, fuego.NotFoundError{
			Detail: "application not found",
		}
	}

	setServersLogCtx := deployRequestData(f.Request(), user)
	appIDStr := data.ApplicationID.String()
	if setServersLogCtx == "" {
		setServersLogCtx = fmt.Sprintf("application_id=%s", appIDStr)
	} else {
		setServersLogCtx = fmt.Sprintf("%s application_id=%s", setServersLogCtx, appIDStr)
	}

	if err := c.storage.SetApplicationServers(data.ApplicationID, data.ServerIDs, data.PrimaryServerID, data.RoutingStrategy); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("deploy: SetApplicationServers: %v", err), setServersLogCtx)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	if reconcileErr := caddy.EnqueueReconcile(organizationID); reconcileErr != nil {
		c.logger.Log(logger.Error, "deploy: failed to enqueue route sync after server change", reconcileErr.Error())
		return nil, fuego.HTTPError{
			Err:    reconcileErr,
			Detail: "failed to enqueue route synchronization: " + reconcileErr.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	servers, err := c.storage.GetApplicationServers(data.ApplicationID)
	if err != nil {
		c.logger.Log(logger.Error, "deploy: failed to reload application servers", err.Error())
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
