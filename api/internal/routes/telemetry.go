package routes

import (
	"github.com/go-fuego/fuego"
	telemetry "github.com/nixopus/nixopus/api/internal/features/telemetry/controller"
)

func (router *Router) RegisterTelemetryRoutes(group *fuego.Server, controller *telemetry.TelemetryController) {
	fuego.Post(group, "", controller.HandleTrackInstall, fuego.OptionSummary("Track CLI installation event"))
}
