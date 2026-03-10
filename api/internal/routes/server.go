package routes

import (
	"github.com/go-fuego/fuego"
	server_controller "github.com/raghavyuva/nixopus-api/internal/features/server/controller"
)

// RegisterServerRoutes registers server management routes
func (router *Router) RegisterServerRoutes(serverGroup *fuego.Server, serverController *server_controller.ServerController) {
	fuego.Get(
		serverGroup,
		"",
		serverController.ListServers,
		fuego.OptionSummary("List servers"),
		fuego.OptionQueryInt("page", "Page number"),
		fuego.OptionQueryInt("page_size", "Page size"),
		fuego.OptionQuery("search", "Search servers by name"),
		fuego.OptionQuery("sort_by", "Sort field"),
		fuego.OptionQuery("sort_order", "Sort order"),
		fuego.OptionQuery("status", "Server status filter"),
		fuego.OptionQueryBool("is_active", "Filter by active state"),
	)
	fuego.Get(
		serverGroup,
		"/ssh/status",
		serverController.CheckSSHStatus,
		fuego.OptionSummary("Get SSH connection status"),
	)
}
