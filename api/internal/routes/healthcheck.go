package routes

import (
	"github.com/go-fuego/fuego"
	healthcheckController "github.com/nixopus/nixopus/api/internal/features/healthcheck/controller"
)

func (router *Router) RegisterHealthCheckRoutes(
	group *fuego.Server,
	controller *healthcheckController.HealthCheckController,
) {
	fuego.Post(group, "", controller.CreateHealthCheck, fuego.OptionSummary("Create health check"))
	fuego.Get(
		group,
		"",
		controller.GetHealthCheck,
		fuego.OptionSummary("Get health checks"),
		fuego.OptionQuery("application_id", "Application ID", fuego.ParamRequired()),
	)
	fuego.Put(group, "", controller.UpdateHealthCheck, fuego.OptionSummary("Update health check"))
	fuego.Delete(
		group,
		"",
		controller.DeleteHealthCheck,
		fuego.OptionSummary("Delete health check"),
		fuego.OptionQuery("application_id", "Application ID", fuego.ParamRequired()),
	)
	fuego.Patch(group, "/toggle", controller.ToggleHealthCheck, fuego.OptionSummary("Toggle health check"))
	fuego.Get(
		group,
		"/results",
		controller.GetHealthCheckResults,
		fuego.OptionSummary("List health check results"),
		fuego.OptionQuery("application_id", "Application ID", fuego.ParamRequired()),
		fuego.OptionQueryInt("limit", "Maximum results"),
		fuego.OptionQuery("start_time", "Start time (RFC3339)"),
		fuego.OptionQuery("end_time", "End time (RFC3339)"),
	)
	fuego.Get(
		group,
		"/stats",
		controller.GetHealthCheckStats,
		fuego.OptionSummary("Get health check stats"),
		fuego.OptionQuery("application_id", "Application ID", fuego.ParamRequired()),
		fuego.OptionQuery("period", "Aggregation period (1h,24h,7d,30d)"),
	)
}
