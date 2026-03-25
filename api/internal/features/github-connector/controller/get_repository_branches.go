package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

type GetGithubRepositoryBranchesRequest struct {
	RepositoryName string `json:"repository_name" validate:"required"`
}

func (c *GithubConnectorController) GetGithubRepositoryBranches(f fuego.ContextWithBody[GetGithubRepositoryBranchesRequest]) (*types.ListBranchesResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	body, err := f.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if strings.TrimSpace(body.RepositoryName) == "" {
		return nil, fuego.BadRequestError{Detail: "repository_name is required", Err: fmt.Errorf("repository_name is required")}
	}

	branches, err := c.service.GetGithubRepositoryBranches(user.ID.String(), body.RepositoryName)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ListBranchesResponse{
		Status:  "success",
		Message: "Branches fetched successfully",
		Data:    branches,
	}, nil
}
