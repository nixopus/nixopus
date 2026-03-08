package routes

import (
	"github.com/go-fuego/fuego"
	trail "github.com/raghavyuva/nixopus-api/internal/features/trail/controller"
)

func (router *Router) RegisterTrailRoutes(group *fuego.Server, controller *trail.TrailController) {
	fuego.Post(group, "/provision", controller.ProvisionTrail)
	fuego.Get(group, "/status/{sessionId}", controller.GetStatus)
}

func (router *Router) RegisterTrailInternalRoutes(group *fuego.Server, controller *trail.TrailController) {
	fuego.Post(group, "/upgrade-resources", controller.UpgradeResources)
}
