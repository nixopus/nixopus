package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *GithubConnectorController) GetGithubRepositories(f fuego.ContextNoBody) (*types.ListRepositoriesResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	q := r.URL.Query()
	page := 1
	pageSize := 10
	connectorID := q.Get("connector_id")
	search := q.Get("search")
	sortBy := q.Get("sort_by")
	sortDirection := q.Get("sort_direction")

	if v := q.Get("page"); v != "" {
		if p, err := strconv.Atoi(v); err == nil && p > 0 {
			page = p
		}
	}
	if v := q.Get("page_size"); v != "" {
		if ps, err := strconv.Atoi(v); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	if sortDirection != "" && sortDirection != "asc" && sortDirection != "desc" {
		sortDirection = "asc" // Default to asc if invalid
	}

	ctxStr := fmt.Sprintf("user_id=%s page=%d page_size=%d connector_id=%s search=%q sort_by=%s sort_dir=%s", user.ID, page, pageSize, connectorID, search, sortBy, sortDirection)
	c.logger.Log(logger.Info, "github connector: GetGithubRepositories", ctxStr)

	repositories, totalCount, err := c.service.GetRepositoriesPaginated(user.ID.String(), page, pageSize, connectorID, search, sortBy, sortDirection)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("github connector: GetGithubRepositories: %v", err), ctxStr)

		errMsg := err.Error()
		if strings.Contains(errMsg, "invalid GitHub installation") || strings.Contains(errMsg, "installation not found") {
			return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
		}
		if strings.Contains(errMsg, "no connector") || strings.Contains(errMsg, "connector not found") {
			return nil, fuego.NotFoundError{Detail: err.Error(), Err: err}
		}
		if strings.Contains(errMsg, "authentication failed") {
			return nil, fuego.UnauthorizedError{Detail: err.Error(), Err: err}
		}

		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "github connector: GetGithubRepositories ok", fmt.Sprintf("%s total=%d count=%d", ctxStr, totalCount, len(repositories)))
	return &types.ListRepositoriesResponse{
		Status:  "success",
		Message: "Repositories fetched successfully",
		Data: types.ListRepositoriesResponseData{
			Repositories: repositories,
			TotalCount:   totalCount,
			Page:         page,
			PageSize:     pageSize,
		},
	}, nil
}
