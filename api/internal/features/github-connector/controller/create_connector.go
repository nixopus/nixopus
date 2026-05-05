package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/utils"

	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (c *GithubConnectorController) CreateGithubConnector(f fuego.ContextWithBody[types.CreateGithubConnectorRequest]) (*types.MessageResponse, error) {
	githubConnectorRequest, err := f.Body()

	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("github connector: CreateGithubConnector body: %v", err), "")
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	w, r := f.Response(), f.Request()
	if err := c.parseAndValidate(w, r, &githubConnectorRequest); err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	ctxStr := fmt.Sprintf("user_id=%s", user.ID)
	c.logger.Log(logger.Info, "github connector: CreateGithubConnector", ctxStr)

	err = c.service.CreateConnector(&githubConnectorRequest, user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("github connector: CreateGithubConnector: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "github connector: CreateGithubConnector ok", ctxStr)
	return &types.MessageResponse{
		Status:  "success",
		Message: "connector created",
	}, nil
}
