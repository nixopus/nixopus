package routes

import (
	"github.com/go-fuego/fuego"
	execute "github.com/raghavyuva/nixopus-api/internal/features/execute/controller"
)

// RegisterExecuteRoutes registers command execution routes.
//
// Routes:
//   - POST /api/v1/execute - Execute a whitelisted command
func (router *Router) RegisterExecuteRoutes(group *fuego.Server, controller *execute.ExecuteController) {
	fuego.Post(group, "", controller.Execute)
}
