package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
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
		c.logMCPDebug("DeleteServer", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMCPDebug("DeleteServer", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		c.logMCPDebug("DeleteServer", "invalid body", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if body.ID == "" {
		c.logMCPDebug("DeleteServer", "server ID required", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "server ID is required"}
	}

	if err := c.service.DeleteServer(body.ID, orgID); err != nil {
		if err == service.ErrServerNotFound {
			return nil, fuego.NotFoundError{Detail: err.Error(), Err: err}
		}
		c.logger.Log(logger.Error, fmt.Sprintf("mcp: DeleteServer: %v", err), fmt.Sprintf("org_id=%s user_id=%s server_id=%s", orgID, user.ID, body.ID))
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	return &Response{
		Status:  "success",
		Message: "MCP server deleted successfully",
	}, nil
}
