package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *GithubConnectorController) DeleteGithubConnector(f fuego.ContextWithBody[types.DeleteGithubConnectorRequest]) (*types.MessageResponse, error) {
	deleteRequest, err := f.Body()

	if err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	w, r := f.Response(), f.Request()

	if !c.parseAndValidate(w, r, &deleteRequest) {
		// parseAndValidate already sent the error response, so return nil to prevent duplicate response
		return nil, nil
	}

	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	err = c.service.DeleteConnector(deleteRequest.ID, user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		if err == types.ErrConnectorDoesNotExist {
			return nil, fuego.NotFoundError{Detail: err.Error(), Err: err}
		}
		if err == types.ErrPermissionDenied {
			return nil, fuego.ForbiddenError{Detail: err.Error(), Err: err}
		}
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "Github Connector deleted successfully",
	}, nil
}
