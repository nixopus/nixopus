package controller

import (
	"errors"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// UpdateContainerResources updates the resource limits (memory, swap, CPU) of a running container.
// It validates the resource limits and verifies the container is running before applying the update.
func (c *ContainerController) UpdateContainerResources(f fuego.ContextWithBody[types.UpdateContainerResourcesRequest]) (*types.UpdateContainerResourcesResponse, error) {
	containerID := f.PathParam("container_id")

	if resp, skipped := c.isProtectedContainer(containerID, "update resources"); skipped {
		return &types.UpdateContainerResourcesResponse{
			Status:  resp.Status,
			Message: resp.Message,
			Data: types.UpdateContainerResourcesResponseData{
				ContainerID: containerID,
			},
		}, nil
	}

	body, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, "Failed to parse request body", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	// Validate resource limits before constructing updateConfig
	const minMemoryBytes = 6 * 1024 * 1024 // 6MB minimum
	if body.Memory != 0 && body.Memory < minMemoryBytes {
		c.logger.Log(logger.Error, "Invalid memory limit", containerID)
		return nil, fuego.HTTPError{
			Err:    errors.New("memory must be 0 or >= 6MB"),
			Status: http.StatusBadRequest,
		}
	}

	// Validate MemorySwap: if non-zero, must be >= Memory or -1 (unlimited swap)
	if body.MemorySwap != 0 && body.MemorySwap != -1 {
		if body.MemorySwap < body.Memory {
			c.logger.Log(logger.Error, "Invalid memory swap limit", containerID)
			return nil, fuego.HTTPError{
				Err:    errors.New("memory_swap must be >= memory, 0 (unlimited), or -1 (unlimited swap)"),
				Status: http.StatusBadRequest,
			}
		}
	}

	// Validate CPUShares: must be >= 2 if provided (non-zero)
	if body.CPUShares != 0 && body.CPUShares < 2 {
		c.logger.Log(logger.Error, "Invalid CPU shares", containerID)
		return nil, fuego.HTTPError{
			Err:    errors.New("cpu_shares must be >= 2"),
			Status: http.StatusBadRequest,
		}
	}

	// Verify container state before updating resources
	containerInfo, err := c.dockerService.GetContainerById(containerID)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to inspect container", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	if containerInfo.State != nil && !containerInfo.State.Running {
		c.logger.Log(logger.Error, "Container is not running", containerID)
		return nil, fuego.HTTPError{
			Err:    errors.New("container must be running to update resource limits"),
			Status: http.StatusBadRequest,
		}
	}

	// Build the update config with the new resource limits
	updateConfig := container.UpdateConfig{
		Resources: container.Resources{
			Memory:     body.Memory,
			MemorySwap: body.MemorySwap,
			CPUShares:  body.CPUShares,
		},
	}

	// Update the container resources
	result, err := c.dockerService.UpdateContainerResources(containerID, updateConfig)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to update container resources", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "Container resources updated successfully", containerID)

	return &types.UpdateContainerResourcesResponse{
		Status:  "success",
		Message: "Container resources updated successfully",
		Data: types.UpdateContainerResourcesResponseData{
			ContainerID: containerID,
			Memory:      body.Memory,
			MemorySwap:  body.MemorySwap,
			CPUShares:   body.CPUShares,
			Warnings:    result.Warnings,
		},
	}, nil
}
