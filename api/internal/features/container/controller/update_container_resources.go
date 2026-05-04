package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/container/service"
	"github.com/nixopus/nixopus/api/internal/features/container/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

// UpdateContainerResources updates the resource limits (memory, swap, CPU) of a running container.
// It validates the resource limits and verifies the container is running before applying the update.
func (c *ContainerController) UpdateContainerResources(f fuego.ContextWithBody[types.UpdateContainerResourcesRequest]) (*types.UpdateContainerResourcesResponse, error) {
	containerID := f.PathParam("container_id")
	if _, err := uuid.Parse(containerID); err != nil {
		c.logger.Log(logger.Debug, "container: UpdateContainerResources invalid id", containerID)
		return nil, fuego.BadRequestError{Detail: "container_id must be a valid UUID"}
	}
	ctx := f.Request().Context()

	if resp, skipped := c.isProtectedContainer(ctx, containerID, "update resources"); skipped {
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
		c.logger.Log(logger.Debug, fmt.Sprintf("container: UpdateContainerResources invalid body: %v", err), fmt.Sprintf("container_id=%s", containerID))
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	ctxStr := fmt.Sprintf("container_id=%s memory=%v memory_swap=%v cpu_shares=%v", containerID, body.Memory, body.MemorySwap, body.CPUShares)
	c.logger.Log(logger.Info, "container: UpdateContainerResources", ctxStr)

	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: UpdateContainerResources docker service: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	opts := service.UpdateContainerResourcesOptions{
		ContainerID: containerID,
		Memory:      body.Memory,
		MemorySwap:  body.MemorySwap,
		CPUShares:   body.CPUShares,
	}

	response, err := service.UpdateContainerResources(dockerService, c.logger, opts)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: UpdateContainerResources: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "container: UpdateContainerResources ok", ctxStr)
	return &response, nil
}
