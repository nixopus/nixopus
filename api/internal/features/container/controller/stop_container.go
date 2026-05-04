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

func (c *ContainerController) StopContainer(f fuego.ContextNoBody) (*types.ContainerActionResponse, error) {
	containerID := f.PathParam("container_id")
	if _, err := uuid.Parse(containerID); err != nil {
		c.logger.Log(logger.Debug, "container: StopContainer invalid id", containerID)
		return nil, fuego.BadRequestError{Detail: "container_id must be a valid UUID"}
	}
	ctx := f.Request().Context()

	if resp, skipped := c.isProtectedContainer(ctx, containerID, "stop"); skipped {
		return resp, nil
	}

	_, r := f.Response(), f.Request()
	orgSettings := c.getOrganizationSettings(r)

	timeout := 10
	if orgSettings.ContainerStopTimeout != nil {
		timeout = *orgSettings.ContainerStopTimeout
	}
	ctxStr := fmt.Sprintf("container_id=%s stop_timeout_sec=%d", containerID, timeout)
	c.logger.Log(logger.Info, "container: StopContainer", ctxStr)

	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: StopContainer docker service: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	opts := service.StopContainerOptions{
		ContainerID: containerID,
		Timeout:     &timeout,
	}

	response, err := service.StopContainer(dockerService, c.logger, opts)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: StopContainer: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "container: StopContainer ok", ctxStr)
	return &response, nil
}
