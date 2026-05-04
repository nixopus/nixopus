package controller

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	mcp "github.com/nixopus/nixopus/api/internal/features/mcp"
	"github.com/nixopus/nixopus/api/internal/features/mcp/service"
	"github.com/nixopus/nixopus/api/internal/features/mcp/validation"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *MCPController) CallTool(f fuego.ContextWithBody[validation.CallToolRequest]) (*Response, error) {
	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)
	if user == nil {
		c.logMCPDebug("CallTool", "authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		c.logMCPDebug("CallTool", "organization ID required", fmt.Sprintf("user_id=%s", user.ID))
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	body, err := f.Body()
	if err != nil {
		c.logMCPDebug("CallTool", "invalid body", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if err := validation.ValidateCallToolRequest(&body); err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("mcp: CallTool validation: %v", err), fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	serverID, err := uuid.Parse(body.ServerID)
	if err != nil {
		c.logMCPDebug("CallTool", "invalid server_id", fmt.Sprintf("org_id=%s user_id=%s", orgID, user.ID))
		return nil, fuego.BadRequestError{Detail: "invalid server_id format"}
	}

	srv, err := c.service.GetServerByID(serverID, orgID)
	if err != nil {
		if errors.Is(err, service.ErrServerNotFound) {
			c.logMCPDebug("CallTool", "server not found", fmt.Sprintf("org_id=%s user_id=%s server_id=%s", orgID, user.ID, serverID))
			return nil, fuego.NotFoundError{Detail: err.Error(), Err: err}
		}
		c.logger.Log(logger.Error, fmt.Sprintf("mcp: CallTool GetServer: %v", err), fmt.Sprintf("org_id=%s user_id=%s server_id=%s", orgID, user.ID, serverID))
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	provider := mcp.GetProvider(srv.ProviderID)
	if provider == nil {
		c.logger.Log(logger.Error, "mcp: CallTool unknown provider", fmt.Sprintf("org_id=%s server_id=%s provider_id=%s", orgID, serverID, srv.ProviderID))
		return nil, fuego.HTTPError{Detail: "unknown provider", Status: http.StatusInternalServerError}
	}

	customURL := ""
	if srv.CustomURL != nil {
		customURL = *srv.CustomURL
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	result, err := service.CallToolOnServer(ctx, provider, customURL, srv.Credentials, service.CallToolParams{
		Name:      body.ToolName,
		Arguments: body.Arguments,
	})
	if err != nil {
		c.logger.Log(logger.Warning, fmt.Sprintf("mcp: CallTool execution failed: %v", err), fmt.Sprintf("org_id=%s user_id=%s server_id=%s tool_name=%s", orgID, user.ID, serverID, body.ToolName))
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusBadGateway}
	}

	return &Response{
		Status:  "success",
		Message: "Tool executed",
		Data:    result,
	}, nil
}
