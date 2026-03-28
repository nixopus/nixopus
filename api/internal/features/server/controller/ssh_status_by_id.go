package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/server/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// CheckSSHStatusByID handles GET /api/v1/servers/{id}/ssh/status.
func (c *ServerController) CheckSSHStatusByID(f fuego.ContextNoBody) (*types.SSHConnectionStatusResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	serverID, err := uuid.Parse(f.PathParam("id"))
	if err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid server ID"}
	}

	response, err := c.service.CheckSSHConnectionByServerID(orgID, serverID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), serverID.String())
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	return response, nil
}
