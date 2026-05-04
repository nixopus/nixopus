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

func (c *MCPController) UpdateServer(f fuego.ContextWithBody[validation.UpdateServerRequest]) (*Response, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		c.logMCPDebug("UpdateServer", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMCPDebug("UpdateServer", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		c.logMCPDebug("UpdateServer", "invalid body", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	body.ID = f.PathParam("id")
	if body.ID == "" {
		c.logMCPDebug("UpdateServer", "server ID required", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "server ID is required"}
	}
	if _, err := uuid.Parse(body.ID); err != nil {
		c.logMCPDebug("UpdateServer", "invalid server UUID", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "server ID must be a valid UUID"}
	}

	if err := validation.ValidateUpdateRequest(&body); err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("mcp: UpdateServer validation: %v", err), fmt.Sprintf("org_id=%s user_id=%s server_id=%s", orgID, user.ID, body.ID))
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	server, err := c.service.UpdateServer(&body, orgID)
	if err != nil {
		if err == service.ErrServerNotFound {
			return nil, fuego.NotFoundError{Detail: err.Error(), Err: err}
		}
		c.logger.Log(logger.Error, fmt.Sprintf("mcp: UpdateServer: %v", err), fmt.Sprintf("org_id=%s user_id=%s server_id=%s", orgID, user.ID, body.ID))
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	return &Response{
		Status:  "success",
		Message: "MCP server updated successfully",
		Data:    toResponse(server),
	}, nil
}
