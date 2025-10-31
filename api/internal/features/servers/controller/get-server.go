package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *ServersController) GetServer(f fuego.ContextNoBody) (*shared_types.Response, error) {
	serverID := f.Request().URL.Query().Get("id")

	if serverID == "" {
		return nil, fuego.HTTPError{
			Err:    types.ErrMissingServerID,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	server, err := c.service.GetServer(serverID, user.ID.String())

	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")

		if err == types.ErrServerNotFound {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusNotFound,
			}
		}

		if isPermissionError(err) {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusForbidden,
			}
		}

		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Server retrieved successfully",
		Data:    server,
	}, nil
}
