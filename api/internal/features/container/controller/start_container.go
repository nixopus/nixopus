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

func (c *ContainerController) StartContainer(f fuego.ContextNoBody) (*types.ContainerActionResponse, error) {
	containerID := f.PathParam("container_id")
	if _, err := uuid.Parse(containerID); err != nil {
		c.logger.Log(logger.Debug, "container: StartContainer invalid id", containerID)
		return nil, fuego.BadRequestError{Detail: "container_id must be a valid UUID"}
	}
	ctx := f.Request().Context()
	ctxStr := fmt.Sprintf("container_id=%s", containerID)
	c.logger.Log(logger.Info, "container: StartContainer", ctxStr)

	if resp, skipped := c.isProtectedContainer(ctx, containerID, "start"); skipped {
		return resp, nil
	}

	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: StartContainer docker service: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	opts := service.StartContainerOptions{
		ContainerID: containerID,
	}

	response, err := service.StartContainer(dockerService, c.logger, opts)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: StartContainer: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "container: StartContainer ok", ctxStr)
	return &response, nil
}
