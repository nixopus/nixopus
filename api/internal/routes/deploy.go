package routes

import (
	"github.com/go-fuego/fuego"
	deploy "github.com/nixopus/nixopus/api/internal/features/deploy/controller"
)

// RegisterDeployRoutes registers deployment routes
func (router *Router) RegisterDeployRoutes(deployGroup *fuego.Server, deployController *deploy.DeployController) {
	fuego.Get(
		deployGroup,
		"/applications",
		deployController.GetApplications,
		fuego.OptionSummary("List applications"),
		fuego.OptionQueryInt("page", "Page number"),
		fuego.OptionQueryInt("page_size", "Page size"),
		fuego.OptionQuery("sort_by", "Sort field"),
		fuego.OptionQuery("sort_direction", "Sort direction"),
	)
	deployApplicationGroup := fuego.Group(deployGroup, "/application")
	router.RegisterDeployApplicationRoutes(deployApplicationGroup, deployController)
}

// RegisterDeployApplicationRoutes registers application-specific deployment routes
func (router *Router) RegisterDeployApplicationRoutes(applicationGroup *fuego.Server, deployController *deploy.DeployController) {
	fuego.Post(
		applicationGroup,
		"",
		deployController.HandleDeploy,
		fuego.OptionSummary("Deploy application"),
	)
	fuego.Post(
		applicationGroup,
		"/project",
		deployController.HandleCreateProject,
		fuego.OptionSummary("Create project"),
	)
	fuego.Post(
		applicationGroup,
		"/project/deploy",
		deployController.HandleDeployProject,
		fuego.OptionSummary("Deploy project"),
	)
	fuego.Post(
		applicationGroup,
		"/project/duplicate",
		deployController.HandleDuplicateProject,
		fuego.OptionSummary("Duplicate project"),
	)
	fuego.Post(
		applicationGroup,
		"/project/add-to-family",
		deployController.HandleAddApplicationToFamily,
		fuego.OptionSummary("Add project to family"),
	)
	fuego.Get(
		applicationGroup,
		"/project/family",
		deployController.HandleGetProjectFamily,
		fuego.OptionSummary("List projects in family"),
		fuego.OptionQuery("family_id", "Project family ID", fuego.ParamRequired()),
	)
	fuego.Get(
		applicationGroup,
		"/project/family/environments",
		deployController.HandleGetEnvironmentsInFamily,
		fuego.OptionSummary("List family environments"),
		fuego.OptionQuery("family_id", "Project family ID", fuego.ParamRequired()),
	)
	fuego.Get(
		applicationGroup,
		"",
		deployController.GetApplicationById,
		fuego.OptionSummary("Get application"),
		fuego.OptionQuery("id", "Application ID", fuego.ParamRequired()),
	)
	fuego.Delete(
		applicationGroup,
		"",
		deployController.DeleteApplication,
		fuego.OptionSummary("Delete application"),
	)
	fuego.Put(
		applicationGroup,
		"",
		deployController.UpdateApplication,
		fuego.OptionSummary("Update application"),
	)
	fuego.Post(
		applicationGroup,
		"/redeploy",
		deployController.ReDeployApplication,
		fuego.OptionSummary("Redeploy application"),
	)
	fuego.Get(
		applicationGroup,
		"/deployments/{deployment_id}",
		deployController.GetDeploymentById,
		fuego.OptionSummary("Get deployment"),
	)
	fuego.Post(
		applicationGroup,
		"/rollback",
		deployController.HandleRollback,
		fuego.OptionSummary("Rollback deployment"),
	)
	fuego.Post(
		applicationGroup,
		"/restart",
		deployController.HandleRestart,
		fuego.OptionSummary("Restart deployment"),
	)
	fuego.Post(
		applicationGroup,
		"/cancel-deployment",
		deployController.CancelDeployment,
		fuego.OptionSummary("Cancel deployment"),
	)
	fuego.Get(
		applicationGroup,
		"/logs/{application_id}",
		deployController.GetLogs,
		fuego.OptionSummary("Get application logs"),
		fuego.OptionQueryInt("page", "Page number"),
		fuego.OptionQueryInt("page_size", "Page size"),
		fuego.OptionQuery("level", "Log level filter"),
		fuego.OptionQuery("start_time", "Start time (RFC3339)"),
		fuego.OptionQuery("end_time", "End time (RFC3339)"),
		fuego.OptionQuery("search_term", "Search term"),
	)
	fuego.Get(
		applicationGroup,
		"/deployments/{deployment_id}/logs",
		deployController.GetDeploymentLogs,
		fuego.OptionSummary("Get deployment logs"),
		fuego.OptionQueryInt("page", "Page number"),
		fuego.OptionQueryInt("page_size", "Page size"),
		fuego.OptionQuery("level", "Log level filter"),
		fuego.OptionQuery("start_time", "Start time (RFC3339)"),
		fuego.OptionQuery("end_time", "End time (RFC3339)"),
		fuego.OptionQuery("search_term", "Search term"),
	)
	fuego.Get(
		applicationGroup,
		"/deployments",
		deployController.GetApplicationDeployments,
		fuego.OptionSummary("List application deployments"),
		fuego.OptionQuery("id", "Application ID", fuego.ParamRequired()),
		fuego.OptionQueryInt("page", "Page number"),
		fuego.OptionQueryInt("limit", "Page size"),
	)
	fuego.Put(
		applicationGroup,
		"/labels",
		deployController.UpdateApplicationLabels,
		fuego.OptionSummary("Update application labels"),
		fuego.OptionQuery("id", "Application ID", fuego.ParamRequired()),
	)
	fuego.Post(
		applicationGroup,
		"/domains",
		deployController.AddApplicationDomain,
		fuego.OptionSummary("Add application domain"),
		fuego.OptionQuery("id", "Application ID", fuego.ParamRequired()),
	)
	fuego.Delete(
		applicationGroup,
		"/domains",
		deployController.RemoveApplicationDomain,
		fuego.OptionSummary("Remove application domain"),
		fuego.OptionQuery("id", "Application ID", fuego.ParamRequired()),
	)
	fuego.Get(
		applicationGroup,
		"/compose-services",
		deployController.GetComposeServices,
		fuego.OptionSummary("List compose services"),
		fuego.OptionQuery("id", "Application ID", fuego.ParamRequired()),
	)
	fuego.Post(
		applicationGroup,
		"/preview-compose",
		deployController.PreviewComposeServices,
		fuego.OptionSummary("Preview compose services"),
	)
	fuego.Post(
		applicationGroup,
		"/recover",
		deployController.HandleRecover,
		fuego.OptionSummary("Recover application"),
	)
	fuego.Get(
		applicationGroup,
		"/servers",
		deployController.GetApplicationServers,
		fuego.OptionSummary("Get application servers"),
		fuego.OptionQuery("id", "Application ID", fuego.ParamRequired()),
	)
	fuego.Put(
		applicationGroup,
		"/servers",
		deployController.SetApplicationServers,
		fuego.OptionSummary("Set application servers"),
	)
}
