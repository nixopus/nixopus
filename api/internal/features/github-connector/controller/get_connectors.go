package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *GithubConnectorController) GetGithubConnectors(f fuego.ContextNoBody) (*types.ListConnectorsResponse, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	ctxStr := fmt.Sprintf("user_id=%s", user.ID)
	c.logger.Log(logger.Info, "github connector: GetGithubConnectors", ctxStr)

	connectors, err := c.service.GetAllConnectors(user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("github connector: GetGithubConnectors: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "github connector: GetGithubConnectors ok", fmt.Sprintf("%s count=%d", ctxStr, len(connectors)))
	return &types.ListConnectorsResponse{
		Status:  "success",
		Message: "Connectors fetched successfully",
		Data:    connectors,
	}, nil
}
