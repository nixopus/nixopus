package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *GithubConnectorController) GetGithubRepositories(f fuego.ContextNoBody) (*types.ListRepositoriesResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
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

	repositories, totalCount, err := c.service.GetGithubRepositoriesPaginated(user.ID.String(), page, pageSize, connectorID, search, sortBy, sortDirection)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		
		errMsg := err.Error()
		if strings.Contains(errMsg, "invalid GitHub installation") || strings.Contains(errMsg, "installation not found") {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusBadRequest,
			}
		}
		if strings.Contains(errMsg, "no connector") || strings.Contains(errMsg, "connector not found") {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusNotFound,
			}
		}
		if strings.Contains(errMsg, "authentication failed") {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusUnauthorized,
			}
		}
		
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

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
