package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/container/service"
	"github.com/nixopus/nixopus/api/internal/features/container/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (c *ContainerController) RemoveContainer(f fuego.ContextNoBody) (*types.ContainerActionResponse, error) {
	containerID := f.PathParam("container_id")
	ctx := f.Request().Context()
	ctxStr := fmt.Sprintf("container_id=%s", containerID)
	c.logger.Log(logger.Info, "container: RemoveContainer", ctxStr)

	if resp, skipped := c.isProtectedContainer(ctx, containerID, "remove"); skipped {
		return resp, nil
	}

	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: RemoveContainer docker service: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	opts := service.RemoveContainerOptions{
		ContainerID: containerID,
		Force:       true,
	}

	response, err := service.RemoveContainer(dockerService, c.logger, opts)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: RemoveContainer: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "container: RemoveContainer ok", ctxStr)
	return &response, nil
}
