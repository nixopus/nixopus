package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *MCPController) ListServers(f fuego.ContextNoBody) (*Response, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)

	servers, err := c.service.ListServers(orgID, false)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	responses := make([]*MCPServerResponse, len(servers))
	for i := range servers {
		responses[i] = toResponse(&servers[i])
	}

	return &Response{
		Status:  "success",
		Message: "MCP servers fetched successfully",
		Data:    responses,
	}, nil
}

func (c *MCPController) ListServersInternal(f fuego.ContextNoBody) (*Response, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)

	servers, err := c.service.ListServers(orgID, true)
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
