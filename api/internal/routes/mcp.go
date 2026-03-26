package routes

import (
	"github.com/go-fuego/fuego"
	mcpController "github.com/nixopus/nixopus/api/internal/features/mcp/controller"
)

// RegisterMCPPublicRoutes registers MCP routes that don't require authentication (e.g. provider icons).
func (router *Router) RegisterMCPPublicRoutes(publicServer *fuego.Server, controller *mcpController.MCPController) {
	iconGroup := fuego.Group(publicServer, "/catalog")
	fuego.Get(iconGroup, "/{provider_id}/icon", controller.GetProviderIcon, fuego.OptionSummary("Get provider icon"))
}

func (router *Router) RegisterMCPRoutes(mcpGroup *fuego.Server, controller *mcpController.MCPController) {
	catalogGroup := fuego.Group(mcpGroup, "/catalog")
	fuego.Get(catalogGroup, "", controller.ListCatalog, fuego.OptionSummary("List MCP provider catalog"))

	serversGroup := fuego.Group(mcpGroup, "/servers")
	fuego.Get(serversGroup, "", controller.ListServers, fuego.OptionSummary("List org MCP servers"))
	fuego.Post(serversGroup, "", controller.AddServer, fuego.OptionSummary("Add MCP server"))
	fuego.Put(serversGroup, "/{id}", controller.UpdateServer, fuego.OptionSummary("Update MCP server"))
	fuego.Delete(serversGroup, "", controller.DeleteServer, fuego.OptionSummary("Delete MCP server"))
	fuego.Post(serversGroup, "/test", controller.TestServer, fuego.OptionSummary("Test MCP server connection"))

	internalGroup := fuego.Group(mcpGroup, "/internal")
	fuego.Get(internalGroup, "/servers", controller.ListServersInternal, fuego.OptionSummary("Agent: list enabled servers with credentials"))
	fuego.Get(internalGroup, "/tools", controller.ListTools, fuego.OptionSummary("Agent: discover tools from all enabled MCP servers"))
}
