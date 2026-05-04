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
		c.logger.Log(logger.Debug, fmt.Sprintf("github connector: GetGithubRepositoryBranches body: %v", err), "")
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if strings.TrimSpace(body.RepositoryName) == "" {
		return nil, fuego.BadRequestError{Detail: "repository_name is required", Err: fmt.Errorf("repository_name is required")}
	}

	ctxStr := fmt.Sprintf("user_id=%s repository=%s", user.ID, body.RepositoryName)
	c.logger.Log(logger.Info, "github connector: GetGithubRepositoryBranches", ctxStr)

	branches, err := c.service.GetRepositoryBranches(user.ID.String(), body.RepositoryName)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("github connector: GetGithubRepositoryBranches: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "github connector: GetGithubRepositoryBranches ok", fmt.Sprintf("%s count=%d", ctxStr, len(branches)))
	return &types.ListBranchesResponse{
		Status:  "success",
		Message: "Branches fetched successfully",
		Data:    branches,
	}, nil
}
