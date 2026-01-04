package controller

import (
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (c *ContainerController) RestartContainer(f fuego.ContextNoBody) (*types.ContainerActionResponse, error) {
	containerID := f.PathParam("container_id")

	if resp, skipped := c.isProtectedContainer(containerID, "restart"); skipped {
		return resp, nil
	}

	_, r := f.Response(), f.Request()
	orgSettings := c.getOrganizationSettings(r)

	// Use timeout from settings, default to 10 seconds if not set
	timeout := 10
	if orgSettings.ContainerStopTimeout != nil {
		timeout = *orgSettings.ContainerStopTimeout
	}

	err := c.dockerService.RestartContainer(containerID, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ContainerActionResponse{
		Status:  "success",
		Message: "Container restarted successfully",
		Data:    types.ContainerStatusData{Status: "restarted"},
	}, nil
}
