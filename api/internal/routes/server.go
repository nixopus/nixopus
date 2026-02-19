package routes

import (
	"github.com/go-fuego/fuego"
	server_controller "github.com/raghavyuva/nixopus-api/internal/features/server/controller"
)

// RegisterServerRoutes registers server management routes
func (router *Router) RegisterServerRoutes(serverGroup *fuego.Server, serverController *server_controller.ServerController) {
	fuego.Get(serverGroup, "", serverController.ListServers)
}
