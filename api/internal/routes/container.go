package routes

import (
	"github.com/go-fuego/fuego"
	container "github.com/nixopus/nixopus/api/internal/features/container/controller"
)

// RegisterContainerRoutes registers container management routes
func (router *Router) RegisterContainerRoutes(containerGroup *fuego.Server, containerController *container.ContainerController) {
	fuego.Get(containerGroup, "", containerController.ListContainers, fuego.OptionSummary("List containers"))
	fuego.Get(containerGroup, "/{container_id}", containerController.GetContainer, fuego.OptionSummary("Get container"))
	fuego.Delete(containerGroup, "/{container_id}", containerController.RemoveContainer, fuego.OptionSummary("Remove container"))
	fuego.Post(containerGroup, "/{container_id}/start", containerController.StartContainer, fuego.OptionSummary("Start container"))
	fuego.Post(containerGroup, "/{container_id}/stop", containerController.StopContainer, fuego.OptionSummary("Stop container"))
	fuego.Post(containerGroup, "/{container_id}/restart", containerController.RestartContainer, fuego.OptionSummary("Restart container"))
	fuego.Post(containerGroup, "/{container_id}/logs", containerController.GetContainerLogs, fuego.OptionSummary("Get container logs"))
	fuego.Put(containerGroup, "/{container_id}/resources", containerController.UpdateContainerResources, fuego.OptionSummary("Update container resources"))
	fuego.Post(containerGroup, "/prune/build-cache", containerController.PruneBuildCache, fuego.OptionSummary("Prune build cache"))
	fuego.Post(containerGroup, "/prune/images", containerController.PruneImages, fuego.OptionSummary("Prune images"))
	fuego.Post(containerGroup, "/images", containerController.ListImages, fuego.OptionSummary("List images"))
}
