package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
)

func (c *ContainerController) StopContainer(f fuego.ContextNoBody) (*types.ContainerActionResponse, error) {
	containerID := f.PathParam("container_id")
	ctx := f.Request().Context()

	if resp, skipped := c.isProtectedContainer(ctx, containerID, "stop"); skipped {
		return resp, nil
	}

	_, r := f.Response(), f.Request()
	orgSettings := c.getOrganizationSettings(r)

	// Use timeout from settings, default to 10 seconds if not set
	timeout := 10
	if orgSettings.ContainerStopTimeout != nil {
		timeout = *orgSettings.ContainerStopTimeout
	}

	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	opts := service.StopContainerOptions{
		ContainerID: containerID,
		Timeout:     &timeout,
	}

	response, err := service.StopContainer(dockerService, c.logger, opts)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &response, nil
}
