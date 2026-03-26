package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/mcp/service"
	"github.com/nixopus/nixopus/api/internal/features/mcp/validation"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *MCPController) AddServer(f fuego.ContextWithBody[validation.CreateServerRequest]) (*Response, error) {
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

	if err := validation.ValidateCreateRequest(&body); err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	server, err := c.service.AddServer(&body, orgID, user.ID)
	if err != nil {
		if err == service.ErrDuplicateName {
			return nil, fuego.ConflictError{Detail: err.Error(), Err: err}
		}
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	return &Response{
		Status:  "success",
		Message: "MCP server added successfully",
		Data:    toResponse(server),
	}, nil
}
