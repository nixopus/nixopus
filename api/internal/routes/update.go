package routes

import (
	"github.com/go-fuego/fuego"
	update "github.com/raghavyuva/nixopus-api/internal/features/update/controller"
)

// RegisterUpdateRoutes registers update routes
func (router *Router) RegisterUpdateRoutes(updateGroup *fuego.Server, updateController *update.UpdateController) {
	fuego.Get(updateGroup, "/check", updateController.CheckForUpdates, fuego.OptionSummary("Check for updates"))
	fuego.Post(updateGroup, "", updateController.PerformUpdate, fuego.OptionSummary("Perform update"))
}
