package service

import (
	"github.com/docker/docker/api/types/container"
	container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// RemoveContainerOptions contains options for removing a container
type RemoveContainerOptions struct {
	ContainerID string
	Force       bool
}

// RemoveContainer removes a Docker container
func RemoveContainer(
	dockerService *docker.DockerService,
	l logger.Logger,
	opts RemoveContainerOptions,
) (container_types.ContainerActionResponse, error) {
	err := dockerService.RemoveContainer(opts.ContainerID, container.RemoveOptions{
		Force: opts.Force,
	})
	if err != nil {
		l.Log(logger.Error, err.Error(), "")
		return container_types.ContainerActionResponse{}, err
	}

	return container_types.ContainerActionResponse{
		Status:  "success",
		Message: "Container removed successfully",
		Data:    container_types.ContainerStatusData{Status: "removed"},
	}, nil
}
