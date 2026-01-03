package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	extension_service "github.com/raghavyuva/nixopus-api/internal/features/extension/service"
	extension_storage "github.com/raghavyuva/nixopus-api/internal/features/extension/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// ListExtensionsHandler returns the handler function for listing extensions
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func ListExtensionsHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, ListExtensionsInput) (*mcp.CallToolResult, ListExtensionsOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input ListExtensionsInput,
	) (*mcp.CallToolResult, ListExtensionsOutput, error) {
		storage := extension_storage.ExtensionStorage{DB: store.DB, Ctx: ctx}
		service := extension_service.NewExtensionService(store, ctx, l, &storage)

		params := shared_types.ExtensionListParams{}

		if input.Category != "" {
			cat := shared_types.ExtensionCategory(input.Category)
			params.Category = &cat
		}

		if input.Type != "" {
			et := shared_types.ExtensionType(input.Type)
			params.Type = &et
		}

		if input.Search != "" {
			params.Search = input.Search
		}

		if input.SortBy != "" {
			params.SortBy = shared_types.ExtensionSortField(input.SortBy)
		}

		if input.SortDir != "" {
			params.SortDir = shared_types.SortDirection(input.SortDir)
		}

		if input.Page != nil {
			params.Page = *input.Page
		}

		if input.PageSize != nil {
			params.PageSize = *input.PageSize
		}

		response, err := service.ListExtensions(params)
		if err != nil {
			return nil, ListExtensionsOutput{}, err
		}

		// Convert shared types to MCP types to avoid circular references
		mcpResponse := convertToMCPExtensionListResponse(response)

		return nil, ListExtensionsOutput{
			Response: mcpResponse,
		}, nil
	}
}

// convertToMCPExtensionListResponse converts shared_types.ExtensionListResponse to MCPExtensionListResponse
// to avoid circular references in the MCP schema
func convertToMCPExtensionListResponse(resp *shared_types.ExtensionListResponse) MCPExtensionListResponse {
	mcpExtensions := make([]MCPExtension, len(resp.Extensions))
	for i, ext := range resp.Extensions {
		mcpExtensions[i] = convertToMCPExtension(ext)
	}

	return MCPExtensionListResponse{
		Extensions: mcpExtensions,
		Total:      resp.Total,
		Page:       resp.Page,
		PageSize:   resp.PageSize,
		TotalPages: resp.TotalPages,
	}
}

// convertToMCPExtension converts shared_types.Extension to MCPExtension
// removing the circular Extension reference from Variables
func convertToMCPExtension(ext shared_types.Extension) MCPExtension {
	mcpVars := make([]MCPExtensionVariable, len(ext.Variables))
	for i, v := range ext.Variables {
		mcpVars[i] = MCPExtensionVariable{
			ID:                v.ID,
			ExtensionID:       v.ExtensionID,
			VariableName:      v.VariableName,
			VariableType:      v.VariableType,
			Description:       v.Description,
			DefaultValue:      v.DefaultValue,
			IsRequired:        v.IsRequired,
			ValidationPattern: v.ValidationPattern,
			CreatedAt:         v.CreatedAt,
		}
	}

	return MCPExtension{
		ID:                ext.ID,
		ExtensionID:       ext.ExtensionID,
		ParentExtensionID: ext.ParentExtensionID,
		Name:              ext.Name,
		Description:       ext.Description,
		Author:            ext.Author,
		Icon:              ext.Icon,
		Category:          ext.Category,
		ExtensionType:     ext.ExtensionType,
		Version:           ext.Version,
		IsVerified:        ext.IsVerified,
		YAMLContent:       ext.YAMLContent,
		ParsedContent:     ext.ParsedContent,
		ContentHash:       ext.ContentHash,
		ValidationStatus:  ext.ValidationStatus,
		ValidationErrors:  ext.ValidationErrors,
		CreatedAt:         ext.CreatedAt,
		UpdatedAt:         ext.UpdatedAt,
		DeletedAt:         ext.DeletedAt,
		Variables:         mcpVars,
	}
}
