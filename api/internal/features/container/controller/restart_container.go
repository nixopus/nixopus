package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
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

	opts := service.RestartContainerOptions{
		ContainerID: containerID,
		Timeout:     &timeout,
	}

	response, err := service.RestartContainer(c.dockerService, c.logger, opts)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &response, nil
}
