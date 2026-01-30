package controller

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/server/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *ServerController) ListServers(f fuego.ContextNoBody) (*types.ListServersResponse, error) {
	w, r := f.Response(), f.Request()
	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	// Get organization ID from context
	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logger.Log(logger.Error, "Organization ID not found in context", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	// Parse query parameters
	params := types.ServerListParams{}

	// Page
	if pageParam := f.QueryParam("page"); pageParam != "" {
		if page, err := strconv.Atoi(pageParam); err == nil && page > 0 {
			params.Page = page
		}
	}

	// Page size
	if pageSizeParam := f.QueryParam("page_size"); pageSizeParam != "" {
		if pageSize, err := strconv.Atoi(pageSizeParam); err == nil && pageSize > 0 {
			params.PageSize = pageSize
		}
	}

	// Search
	if searchParam := f.QueryParam("search"); searchParam != "" {
		params.Search = strings.TrimSpace(searchParam)
	}

	// Sort by
	if sortByParam := f.QueryParam("sort_by"); sortByParam != "" {
		params.SortBy = strings.ToLower(strings.TrimSpace(sortByParam))
	}

	// Sort order
	if sortOrderParam := f.QueryParam("sort_order"); sortOrderParam != "" {
		params.SortOrder = strings.ToLower(strings.TrimSpace(sortOrderParam))
	}

	// Status filter
	if statusParam := f.QueryParam("status"); statusParam != "" {
		params.Status = strings.TrimSpace(statusParam)
	}

	// Is active filter
	if isActiveParam := f.QueryParam("is_active"); isActiveParam != "" {
		if isActive, err := strconv.ParseBool(isActiveParam); err == nil {
			params.IsActive = &isActive
		}
	}

	// Call service layer
	response, err := c.service.ListServers(orgID, params)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return response, nil
}
