package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/server/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *ServerController) CheckSSHStatus(f fuego.ContextNoBody) (*types.SSHConnectionStatusResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	// Get organization ID from context
	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logger.Log(logger.Error, "Organization ID not found in context", "")
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	// Call service layer
	response, err := c.service.CheckSSHConnection(orgID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return response, nil
}
