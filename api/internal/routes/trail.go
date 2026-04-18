package routes

import (
	"github.com/go-fuego/fuego"
	machine_controller "github.com/nixopus/nixopus/api/internal/features/machine/controller"
)

func (router *Router) RegisterTrailRoutes(group *fuego.Server, controller *machine_controller.TrailController) {
	fuego.Post(group, "/provision", controller.ProvisionTrail, fuego.OptionSummary("Provision trail resources"))
	fuego.Get(group, "/status/{sessionId}", controller.GetStatus, fuego.OptionSummary("Get trail session status"))
}

func (router *Router) RegisterTrailInternalRoutes(group *fuego.Server, controller *machine_controller.TrailController) {
	fuego.Post(
		group,
		"/upgrade-resources",
		controller.UpgradeResources,
		fuego.OptionSummary("Upgrade trail resources"),
		fuego.OptionHide(),
	)
}
