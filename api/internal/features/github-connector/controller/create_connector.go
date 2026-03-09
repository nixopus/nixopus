package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/utils"

	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (c *GithubConnectorController) CreateGithubConnector(f fuego.ContextWithBody[types.CreateGithubConnectorRequest]) (*types.MessageResponse, error) {
	githubConnectorRequest, err := f.Body()

	if err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	w, r := f.Response(), f.Request()
	if !c.parseAndValidate(w, r, &githubConnectorRequest) {
		return nil, fuego.BadRequestError{Detail: "invalid request"}
	}

	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	err = c.service.CreateConnector(&githubConnectorRequest, user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "connector created",
	}, nil
}
