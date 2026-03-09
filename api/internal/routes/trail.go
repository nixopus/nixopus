package routes

import (
	"github.com/go-fuego/fuego"
	trail "github.com/nixopus/nixopus/api/internal/features/trail/controller"
)

func (router *Router) RegisterTrailRoutes(group *fuego.Server, controller *trail.TrailController) {
	fuego.Post(group, "/provision", controller.ProvisionTrail, fuego.OptionSummary("Provision trail resources"))
	fuego.Get(group, "/status/{sessionId}", controller.GetStatus, fuego.OptionSummary("Get trail session status"))
}

func (router *Router) RegisterTrailInternalRoutes(group *fuego.Server, controller *trail.TrailController) {
	fuego.Post(
		group,
		"/upgrade-resources",
		controller.UpgradeResources,
		fuego.OptionSummary("Upgrade trail resources"),
		fuego.OptionHide(),
	)
}
