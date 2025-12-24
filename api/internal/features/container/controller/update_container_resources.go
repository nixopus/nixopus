package controller

import (
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// UpdateContainerResources updates the resource limits (memory, swap, CPU) of a container
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
