package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *GithubConnectorController) DeleteGithubConnector(f fuego.ContextWithBody[types.DeleteGithubConnectorRequest]) (*types.MessageResponse, error) {
	deleteRequest, err := f.Body()

	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("github connector: DeleteGithubConnector body: %v", err), "")
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	w, r := f.Response(), f.Request()

	if err := c.parseAndValidate(w, r, &deleteRequest); err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	ctxStr := fmt.Sprintf("user_id=%s connector_id=%s", user.ID, deleteRequest.ID)
	c.logger.Log(logger.Info, "github connector: DeleteGithubConnector", ctxStr)

	err = c.service.DeleteConnector(deleteRequest.ID, user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("github connector: DeleteGithubConnector: %v", err), ctxStr)
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

	c.logger.Log(logger.Info, "github connector: DeleteGithubConnector ok", ctxStr)
	return &types.MessageResponse{
		Status:  "success",
		Message: "Github Connector deleted successfully",
	}, nil
}
