package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
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
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	opts := service.UpdateContainerResourcesOptions{
		ContainerID: containerID,
		Memory:      body.Memory,
		MemorySwap:  body.MemorySwap,
		CPUShares:   body.CPUShares,
	}

	response, err := service.UpdateContainerResources(c.dockerService, c.logger, opts)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &response, nil
}
