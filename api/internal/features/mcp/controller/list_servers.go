package controller

import (
	"net/http"
	"strconv"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/mcp/storage"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *MCPController) ListServers(f fuego.ContextNoBody) (*Response, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 {
		limit = 20
	}

	params := storage.ListServersParams{
		Q:       q.Get("q"),
		SortBy:  q.Get("sort_by"),
		SortDir: q.Get("sort_dir"),
		Page:    page,
		Limit:   limit,
	}

	servers, totalCount, err := c.service.ListServers(orgID, params)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	items := make([]*MCPServerResponse, len(servers))
	for i := range servers {
		items[i] = toResponse(&servers[i])
	}

	return &Response{
		Status:  "success",
		Message: "MCP servers fetched successfully",
		Data: PaginatedData[*MCPServerResponse]{
			Items:      items,
			TotalCount: totalCount,
			Page:       page,
			PageSize:   limit,
		},
	}, nil
}

func (c *MCPController) ListServersInternal(f fuego.ContextNoBody) (*Response, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)

	servers, _, err := c.service.ListServers(orgID, storage.ListServersParams{EnabledOnly: true})
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	return &Response{
		Status:  "success",
		Message: "MCP servers fetched successfully",
		Data:    servers,
	}, nil
}
