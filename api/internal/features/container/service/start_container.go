package service

import (
	"github.com/docker/docker/api/types/container"
	container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// StartContainerOptions contains options for starting a container
type StartContainerOptions struct {
	ContainerID string
}

// StartContainer starts a Docker container
func StartContainer(
	dockerService *docker.DockerService,
	l logger.Logger,
	opts StartContainerOptions,
) (container_types.ContainerActionResponse, error) {
	err := dockerService.StartContainer(opts.ContainerID, container.StartOptions{})
	if err != nil {
		l.Log(logger.Error, err.Error(), "")
		return container_types.ContainerActionResponse{}, err
	}

	return container_types.ContainerActionResponse{
		Status:  "success",
		Message: "Container started successfully",
		Data:    container_types.ContainerStatusData{Status: "started"},
	}, nil
}
