package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/mcp/service"
	"github.com/nixopus/nixopus/api/internal/features/mcp/validation"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *MCPController) AddServer(f fuego.ContextWithBody[validation.CreateServerRequest]) (*Response, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		c.logMCPDebug("AddServer", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMCPDebug("AddServer", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		c.logMCPDebug("AddServer", "invalid body", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if err := validation.ValidateCreateRequest(&body); err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("mcp: AddServer validation: %v", err), fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	server, err := c.service.AddServer(&body, orgID, user.ID)
	if err != nil {
		if err == service.ErrDuplicateName {
			return nil, fuego.ConflictError{Detail: err.Error(), Err: err}
		}
		c.logger.Log(logger.Error, fmt.Sprintf("mcp: AddServer: %v", err), fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	return &Response{
		Status:  "success",
		Message: "MCP server added successfully",
		Data:    toResponse(server),
	}, nil
}
