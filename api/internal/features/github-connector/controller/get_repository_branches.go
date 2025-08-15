package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type GetGithubRepositoryBranchesRequest struct {
	RepositoryName string `json:"repository_name" validate:"required"`
}

func (c *GithubConnectorController) GetGithubRepositoryBranches(f fuego.ContextWithBody[GetGithubRepositoryBranchesRequest]) (*shared_types.Response, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	body, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	if strings.TrimSpace(body.RepositoryName) == "" {
		return nil, fuego.HTTPError{
			Err:    fmt.Errorf("repository_name is required"),
			Status: http.StatusBadRequest,
		}
	}

	branches, err := c.service.GetGithubRepositoryBranches(user.ID.String(), body.RepositoryName)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Branches fetched successfully",
		Data:    branches,
	}, nil
}
