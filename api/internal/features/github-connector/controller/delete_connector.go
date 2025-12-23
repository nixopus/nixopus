package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *GithubConnectorController) DeleteGithubConnector(f fuego.ContextWithBody[types.DeleteGithubConnectorRequest]) (*shared_types.Response, error) {
	deleteRequest, err := f.Body()

	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := f.Response(), f.Request()

	if !c.parseAndValidate(w, r, &deleteRequest) {
		// parseAndValidate already sent the error response, so return nil to prevent duplicate response
		return nil, nil
	}

	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	err = c.service.DeleteConnector(deleteRequest.ID, user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		if err == types.ErrConnectorDoesNotExist {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusNotFound,
			}
		}
		if err == types.ErrPermissionDenied {
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
		Message: "Github Connector deleted successfully",
	}, nil
}
