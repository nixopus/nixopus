package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *GithubConnectorController) GetGithubConnectors(f fuego.ContextNoBody) (*types.ListConnectorsResponse, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	connectors, err := c.service.GetAllConnectors(user.ID.String())
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ListConnectorsResponse{
		Status:  "success",
		Message: "Connectors fetched successfully",
		Data:    connectors,
	}, nil
}
