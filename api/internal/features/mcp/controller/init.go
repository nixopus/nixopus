package controller

import (
	"context"
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	mcp "github.com/nixopus/nixopus/api/internal/features/mcp"
	"github.com/nixopus/nixopus/api/internal/features/mcp/service"
	"github.com/nixopus/nixopus/api/internal/features/mcp/storage"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

type MCPController struct {
	store   *shared_storage.Store
	service *service.MCPService
	ctx     context.Context
	logger  logger.Logger
}

type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

func NewMCPController(store *shared_storage.Store, ctx context.Context, l logger.Logger) *MCPController {
	repo := storage.MCPStorage{DB: store.DB, Ctx: ctx}
	svc := service.NewMCPService(store, ctx, l, repo)
	return &MCPController{store: store, service: svc, ctx: ctx, logger: l}
}

func resolveURL(server *shared_types.MCPServer) string {
	if server.ProviderID == "custom" && server.CustomURL != nil {
		return *server.CustomURL
	}
	provider := mcp.GetProvider(server.ProviderID)
	if provider == nil {
		return ""
	}
	return provider.URL
}

func maskCredentials(server *shared_types.MCPServer) map[string]string {
	provider := mcp.GetProvider(server.ProviderID)
	masked := make(map[string]string, len(server.Credentials))
	for k, v := range server.Credentials {
		if v == "" {
			masked[k] = ""
			continue
		}
		isSensitive := true
		if provider != nil {
			for _, field := range provider.Fields {
				if field.Key == k {
					isSensitive = field.Sensitive
					break
				}
			}
		}
		if isSensitive {
			masked[k] = "***"
		} else {
			masked[k] = v
		}
	}
	return masked
}

type MCPServerResponse struct {
	*shared_types.MCPServer
	MaskedCredentials map[string]string `json:"credentials"`
	ResolvedURL       string            `json:"url"`
}

func toResponse(server *shared_types.MCPServer) *MCPServerResponse {
	return &MCPServerResponse{
		MCPServer:         server,
		MaskedCredentials: maskCredentials(server),
		ResolvedURL:       resolveURL(server),
	}
}

type catalogEntryWithLogo struct {
	mcp.MCPProvider
	LogoURL string `json:"logo_url"`
}

func withLogoURL(p mcp.MCPProvider) catalogEntryWithLogo {
	return catalogEntryWithLogo{
		MCPProvider: p,
		LogoURL:     fmt.Sprintf("/v1/mcp/catalog/%s/icon", p.ID),
	}
}

type PaginatedData[T any] struct {
	Items      []T `json:"items"`
	TotalCount int `json:"total_count"`
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
}
