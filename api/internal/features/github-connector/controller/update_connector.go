package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/utils"

	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (c *GithubConnectorController) UpdateGithubConnectorRequest(f fuego.ContextWithBody[types.UpdateGithubConnectorRequest]) (*types.MessageResponse, error) {

	UpdateConnectorRequest, err := f.Body()

	if err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	w, r := f.Response(), f.Request()

	if !c.parseAndValidate(w, r, &UpdateConnectorRequest) {
		return nil, fuego.BadRequestError{Detail: "invalid request"}
	}

	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	err = c.service.UpdateGithubConnectorRequest(UpdateConnectorRequest.InstallationID, user.ID.String(), UpdateConnectorRequest.ConnectorID)
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
		Message: "Github Connector Request Updated",
	}, nil
}
