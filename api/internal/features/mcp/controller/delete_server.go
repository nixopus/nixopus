package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/mcp/service"
	"github.com/nixopus/nixopus/api/internal/utils"
)

type DeleteServerRequest struct {
	ID string `json:"id"`
}

func (c *MCPController) DeleteServer(f fuego.ContextWithBody[DeleteServerRequest]) (*Response, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)

	body, err := f.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if body.ID == "" {
		return nil, fuego.BadRequestError{Detail: "server ID is required"}
	}

	if err := c.service.DeleteServer(body.ID, orgID); err != nil {
		if err == service.ErrServerNotFound {
			return nil, fuego.NotFoundError{Detail: err.Error(), Err: err}
		}
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	return &Response{
		Status:  "success",
		Message: "MCP server deleted successfully",
	}, nil
}
