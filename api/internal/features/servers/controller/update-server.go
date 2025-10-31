package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *ServersController) UpdateServer(f fuego.ContextWithBody[types.UpdateServerRequest]) (*shared_types.Response, error) {
	serverRequest, err := f.Body()

	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()
	if !c.parseAndValidate(w, r, &serverRequest) {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	updated, err := c.service.UpdateServer(serverRequest, user.ID.String())

	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")

		if isInvalidServerError(err) {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusBadRequest,
			}
		}

		if err == types.ErrServerNotFound {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusNotFound,
			}
		}

		if err == types.ErrServerAlreadyExists || err == types.ErrServerHostAlreadyExists {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusConflict,
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
		Message: "Server updated successfully",
		Data:    updated,
	}, nil
}
