package controller

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	mcp "github.com/nixopus/nixopus/api/internal/features/mcp"
	"github.com/nixopus/nixopus/api/internal/features/mcp/service"
	"github.com/nixopus/nixopus/api/internal/features/mcp/storage"
	"github.com/nixopus/nixopus/api/internal/utils"
)

// ListTools discovers available tools from all enabled MCP servers for the org.
// Requests to each server are made concurrently with a 20-second timeout.
// Per-server errors are captured in the ErrorMessage field rather than failing the whole request.
func (c *MCPController) ListTools(f fuego.ContextNoBody) (*Response, error) {
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

	ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
	defer cancel()

	results := make([]service.ServerToolSet, len(servers))
	var wg sync.WaitGroup

	for i, srv := range servers {
		i, srv := i, srv
		wg.Add(1)
		go func() {
			defer wg.Done()

			results[i] = service.ServerToolSet{
				ServerID:   srv.ID.String(),
				ServerName: srv.Name,
				ProviderID: srv.ProviderID,
				Tools:      []service.MCPTool{},
			}

			provider := mcp.GetProvider(srv.ProviderID)
			if provider == nil {
				results[i].Error = "unknown provider"
				return
			}

			customURL := ""
			if srv.CustomURL != nil {
				customURL = *srv.CustomURL
			}

			tools, err := service.DiscoverServerTools(ctx, provider, customURL, srv.Credentials)
			if err != nil {
				c.logger.Log(logger.Warning, "tool discovery failed for "+srv.Name, err.Error())
				results[i].Error = err.Error()
				return
			}
			results[i].Tools = tools
		}()
	}

	wg.Wait()

	return &Response{
		Status:  "success",
		Message: "Tool discovery complete",
		Data:    results,
	}, nil
}
