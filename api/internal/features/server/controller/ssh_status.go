package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/server/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *ServerController) CheckSSHStatus(f fuego.ContextNoBody) (*types.SSHConnectionStatusResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	// Get organization ID from context
	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logger.Log(logger.Error, "Organization ID not found in context", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	// Call service layer
	response, err := c.service.CheckSSHConnection(orgID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return response, nil
}
