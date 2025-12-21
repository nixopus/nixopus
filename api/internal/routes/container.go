package routes

import (
	"github.com/go-fuego/fuego"
	container "github.com/raghavyuva/nixopus-api/internal/features/container/controller"
)

// RegisterContainerRoutes registers container management routes
func (router *Router) RegisterContainerRoutes(containerGroup *fuego.Server, containerController *container.ContainerController) {
	fuego.Get(containerGroup, "", containerController.ListContainers)
	fuego.Get(containerGroup, "/{container_id}", containerController.GetContainer)
	fuego.Delete(containerGroup, "/{container_id}", containerController.RemoveContainer)
	fuego.Post(containerGroup, "/{container_id}/start", containerController.StartContainer)
	fuego.Post(containerGroup, "/{container_id}/stop", containerController.StopContainer)
	fuego.Post(containerGroup, "/{container_id}/restart", containerController.RestartContainer)
	fuego.Post(containerGroup, "/{container_id}/logs", containerController.GetContainerLogs)
	fuego.Post(containerGroup, "/prune/build-cache", containerController.PruneBuildCache)
	fuego.Post(containerGroup, "/prune/images", containerController.PruneImages)
	fuego.Post(containerGroup, "/images", containerController.ListImages)
}
