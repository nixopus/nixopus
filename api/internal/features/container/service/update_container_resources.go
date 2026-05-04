package service

import (
	"errors"
	"fmt"

	"github.com/docker/docker/api/types/container"
	container_types "github.com/nixopus/nixopus/api/internal/features/container/types"
	"github.com/nixopus/nixopus/api/internal/features/deploy/docker"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

// UpdateContainerResourcesOptions contains options for updating container resources
type UpdateContainerResourcesOptions struct {
	ContainerID string
	Memory      int64
	MemorySwap  int64
	CPUShares   int64
}

// UpdateContainerResources updates the resource limits of a running container
func UpdateContainerResources(
	dockerService docker.DockerRepository,
	l logger.Logger,
	opts UpdateContainerResourcesOptions,
) (container_types.UpdateContainerResourcesResponse, error) {
	// Validate resource limits before constructing updateConfig
	const minMemoryBytes = 6 * 1024 * 1024 // 6MB minimum
	if opts.Memory != 0 && opts.Memory < minMemoryBytes {
		l.Log(logger.Error, "container service: UpdateContainerResources invalid memory", opts.ContainerID)
		return container_types.UpdateContainerResourcesResponse{}, errors.New("memory must be 0 or >= 6MB")
	}

	// Validate MemorySwap: if non-zero, must be >= Memory or -1 (unlimited swap)
	if opts.MemorySwap != 0 && opts.MemorySwap != -1 {
		if opts.MemorySwap < opts.Memory {
			l.Log(logger.Error, "container service: UpdateContainerResources invalid memory_swap", opts.ContainerID)
			return container_types.UpdateContainerResourcesResponse{}, errors.New("memory_swap must be >= memory, 0 (unlimited), or -1 (unlimited swap)")
		}
	}

	// Validate CPUShares: must be >= 2 if provided (non-zero)
	if opts.CPUShares != 0 && opts.CPUShares < 2 {
		l.Log(logger.Error, "container service: UpdateContainerResources invalid cpu_shares", opts.ContainerID)
		return container_types.UpdateContainerResourcesResponse{}, errors.New("cpu_shares must be >= 2")
	}

	// Verify container state before updating resources
	containerInfo, err := dockerService.GetContainerById(opts.ContainerID)
	if err != nil {
		l.Log(logger.Error, fmt.Sprintf("container service: UpdateContainerResources inspect: %v", err), opts.ContainerID)
		return container_types.UpdateContainerResourcesResponse{}, err
	}

	if containerInfo.State != nil && !containerInfo.State.Running {
		l.Log(logger.Error, "container service: UpdateContainerResources not running", opts.ContainerID)
		return container_types.UpdateContainerResourcesResponse{}, errors.New("container must be running to update resource limits")
	}

	// Build the update config with the new resource limits
	updateConfig := container.UpdateConfig{
		Resources: container.Resources{
			Memory:     opts.Memory,
			MemorySwap: opts.MemorySwap,
			CPUShares:  opts.CPUShares,
		},
	}

	// Update the container resources
	result, err := dockerService.UpdateContainerResources(opts.ContainerID, updateConfig)
	if err != nil {
		l.Log(logger.Error, fmt.Sprintf("container service: UpdateContainerResources docker update: %v", err), opts.ContainerID)
		return container_types.UpdateContainerResourcesResponse{}, err
	}

	l.Log(logger.Info, "container service: UpdateContainerResources ok", opts.ContainerID)

	return container_types.UpdateContainerResourcesResponse{
		Status:  "success",
		Message: "Container resources updated successfully",
		Data: container_types.UpdateContainerResourcesResponseData{
			ContainerID: opts.ContainerID,
			Memory:      opts.Memory,
			MemorySwap:  opts.MemorySwap,
			CPUShares:   opts.CPUShares,
			Warnings:    result.Warnings,
		},
	}, nil
}
