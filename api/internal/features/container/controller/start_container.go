package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
)

func (c *ContainerController) StartContainer(f fuego.ContextNoBody) (*types.ContainerActionResponse, error) {
	containerID := f.PathParam("container_id")
	ctx := f.Request().Context()

	if resp, skipped := c.isProtectedContainer(ctx, containerID, "start"); skipped {
		return resp, nil
	}

	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	opts := service.StartContainerOptions{
		ContainerID: containerID,
	}

	response, err := service.StartContainer(dockerService, c.logger, opts)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &response, nil
}
