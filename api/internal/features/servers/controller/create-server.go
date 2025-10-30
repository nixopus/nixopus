package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *ServersController) CreateServer(f fuego.ContextWithBody[types.CreateServerRequest]) (*shared_types.Response, error) {
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

	organization := utils.GetOrganizationID(r)

	if organization == uuid.Nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	created, err := c.service.CreateServer(serverRequest, user.ID.String(), organization.String())

	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")

		if isInvalidServerError(err) {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusBadRequest,
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
		Message: "Server created successfully",
		Data:    created,
	}, nil
}