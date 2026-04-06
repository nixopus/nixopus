package controller

import (
	"context"
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
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	body, err := f.Body()
	if err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	if err := validation.ValidateCallToolRequest(&body); err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	serverID, err := uuid.Parse(body.ServerID)
	if err != nil {
		return nil, fuego.BadRequestError{Detail: "invalid server_id format"}
	}

	orgID := utils.GetOrganizationID(r)

	srv, err := c.service.GetServerByID(serverID, orgID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.NotFoundError{Detail: "server not found", Err: err}
	}

	provider := mcp.GetProvider(srv.ProviderID)
	if provider == nil {
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
		c.logger.Log(logger.Warning, "tool call failed: "+err.Error(), body.ToolName)
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusBadGateway}
	}

	return &Response{
		Status:  "success",
		Message: "Tool executed",
		Data:    result,
	}, nil
}
