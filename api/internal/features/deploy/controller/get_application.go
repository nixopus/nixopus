package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *DeployController) GetApplicationById(f fuego.ContextNoBody) (*shared_types.Response, error) {
	id := f.QueryParam("id")

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "user not found", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		c.logger.Log(logger.Error, "organization not found", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	// Get server ID from request context (set by auth middleware from X-Server-Id header)
	var serverID *uuid.UUID
	server := utils.GetServer(f.Response(), f.Request())
	if server != nil {
		serverID = &server.ID
	}

	application, err := c.service.GetApplicationById(id, organizationID, serverID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Application Retrieved successfully",
		Data:    application,
	}, nil
}
