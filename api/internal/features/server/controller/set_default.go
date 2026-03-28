package controller

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/server/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// SetDefaultServer handles PUT /api/v1/servers/{id}/set-default.
func (c *ServerController) SetDefaultServer(f fuego.ContextNoBody) (*types.SetDefaultServerResponse, error) {
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

	key, err := c.service.SetDefaultServer(orgID, serverID)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrServerNotFound), errors.Is(err, sql.ErrNoRows):
			return nil, fuego.NotFoundError{Detail: "server not found"}
		case errors.Is(err, types.ErrServerInactive):
			return nil, fuego.BadRequestError{Detail: err.Error()}
		default:
			c.logger.Log(logger.Error, err.Error(), serverID.String())
			return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
		}
	}

	return &types.SetDefaultServerResponse{
		Status:  "success",
		Message: "Server set as default successfully",
		Data:    *key,
	}, nil
}
