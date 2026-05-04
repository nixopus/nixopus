package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/utils"

	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (c *GithubConnectorController) UpdateGithubConnectorRequest(f fuego.ContextWithBody[types.UpdateGithubConnectorRequest]) (*types.MessageResponse, error) {

	UpdateConnectorRequest, err := f.Body()

	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("github connector: UpdateGithubConnector body: %v", err), "")
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	w, r := f.Response(), f.Request()

	if err := c.parseAndValidate(w, r, &UpdateConnectorRequest); err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	ctxStr := fmt.Sprintf("user_id=%s connector_id=%s", user.ID, UpdateConnectorRequest.ConnectorID)
	c.logger.Log(logger.Info, "github connector: UpdateGithubConnector", ctxStr)

	err = c.service.UpdateConnectorInstallation(UpdateConnectorRequest.InstallationID, user.ID.String(), UpdateConnectorRequest.ConnectorID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("github connector: UpdateGithubConnector: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "github connector: UpdateGithubConnector ok", ctxStr)
	return &types.MessageResponse{
		Status:  "success",
		Message: "Github Connector Request Updated",
	}, nil
}
