package service

import (
	"github.com/docker/docker/api/types/container"
	container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// RestartContainerOptions contains options for restarting a container
type RestartContainerOptions struct {
	ContainerID string
	Timeout     *int
}

// RestartContainer restarts a Docker container
func RestartContainer(
	dockerService *docker.DockerService,
	l logger.Logger,
	opts RestartContainerOptions,
) (container_types.ContainerActionResponse, error) {
	stopOpts := container.StopOptions{}
	if opts.Timeout != nil {
		stopOpts.Timeout = opts.Timeout
	}

	err := dockerService.RestartContainer(opts.ContainerID, stopOpts)
	if err != nil {
		l.Log(logger.Error, err.Error(), "")
		return container_types.ContainerActionResponse{}, err
	}

	return container_types.ContainerActionResponse{
		Status:  "success",
		Message: "Container restarted successfully",
		Data:    container_types.ContainerStatusData{Status: "restarted"},
	}, nil
}
