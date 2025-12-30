package routes

import (
	"github.com/go-fuego/fuego"
	healthcheckController "github.com/raghavyuva/nixopus-api/internal/features/healthcheck/controller"
)

func (router *Router) RegisterHealthCheckRoutes(
	group *fuego.Server,
	controller *healthcheckController.HealthCheckController,
) {
	fuego.Post(group, "", controller.CreateHealthCheck)
	fuego.Get(group, "", controller.GetHealthCheck)
	fuego.Put(group, "", controller.UpdateHealthCheck)
	fuego.Delete(group, "", controller.DeleteHealthCheck)
	fuego.Patch(group, "/toggle", controller.ToggleHealthCheck)
	fuego.Get(group, "/results", controller.GetHealthCheckResults)
	fuego.Get(group, "/stats", controller.GetHealthCheckStats)
}
