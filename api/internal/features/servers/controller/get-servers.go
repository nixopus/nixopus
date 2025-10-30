package controller

import (
	"net/http"
	"strconv"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/servers/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *ServersController) GetServers(f fuego.ContextNoBody) (*shared_types.Response, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID := utils.GetOrganizationID(r)
	if organizationID == uuid.Nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	queryParams := parseServerQueryParams(r)
	queryParams.SetDefaults()

	if !queryParams.IsValidSortField() {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	serverListResponse, err := c.service.GetServers(user.ID.String(), organizationID.String(), queryParams)

	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")

		if isPermissionError(err) {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusForbidden,
			}
		}

		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Servers retrieved successfully",
		Data:    serverListResponse,
	}, nil
}

func parseServerQueryParams(r *http.Request) *types.ServerQueryParams {
	params := &types.ServerQueryParams{}

	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			params.Page = page
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			params.PageSize = pageSize
		}
	}

	params.Search = r.URL.Query().Get("search")

	params.SortBy = r.URL.Query().Get("sort_by")

	params.SortOrder = r.URL.Query().Get("sort_order")

	return params
}
