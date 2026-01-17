package routes

import (
	"github.com/go-fuego/fuego"
	deploy "github.com/raghavyuva/nixopus-api/internal/features/deploy/controller"
)

// RegisterDeployRoutes registers deployment routes
func (router *Router) RegisterDeployRoutes(deployGroup *fuego.Server, deployController *deploy.DeployController) {
	fuego.Get(deployGroup, "/applications", deployController.GetApplications)
	deployApplicationGroup := fuego.Group(deployGroup, "/application")
	router.RegisterDeployApplicationRoutes(deployApplicationGroup, deployController)
}

// RegisterDeployApplicationRoutes registers application-specific deployment routes
func (router *Router) RegisterDeployApplicationRoutes(applicationGroup *fuego.Server, deployController *deploy.DeployController) {
	fuego.Post(applicationGroup, "", deployController.HandleDeploy)
	fuego.Post(applicationGroup, "/project", deployController.HandleCreateProject)
	fuego.Post(applicationGroup, "/project/deploy", deployController.HandleDeployProject)
	fuego.Post(applicationGroup, "/project/duplicate", deployController.HandleDuplicateProject)
	fuego.Get(applicationGroup, "/project/family", deployController.HandleGetProjectFamily)
	fuego.Get(applicationGroup, "/project/family/environments", deployController.HandleGetEnvironmentsInFamily)
	fuego.Get(applicationGroup, "", deployController.GetApplicationById)
	fuego.Delete(applicationGroup, "", deployController.DeleteApplication)
	fuego.Put(applicationGroup, "", deployController.UpdateApplication)
	fuego.Post(applicationGroup, "/redeploy", deployController.ReDeployApplication)
	fuego.Get(applicationGroup, "/deployments/{deployment_id}", deployController.GetDeploymentById)
	fuego.Post(applicationGroup, "/rollback", deployController.HandleRollback)
	fuego.Post(applicationGroup, "/restart", deployController.HandleRestart)
	fuego.Get(applicationGroup, "/logs/{application_id}", deployController.GetLogs)
	fuego.Get(applicationGroup, "/deployments/{deployment_id}/logs", deployController.GetDeploymentLogs)
	fuego.Get(applicationGroup, "/deployments", deployController.GetApplicationDeployments)
	fuego.Put(applicationGroup, "/labels", deployController.UpdateApplicationLabels)
	fuego.Post(applicationGroup, "/domains", deployController.AddApplicationDomain)
	fuego.Delete(applicationGroup, "/domains", deployController.RemoveApplicationDomain)
}
