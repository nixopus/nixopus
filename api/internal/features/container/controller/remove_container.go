package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
)

func (c *ContainerController) RemoveContainer(f fuego.ContextNoBody) (*types.ContainerActionResponse, error) {
	containerID := f.PathParam("container_id")
	ctx := f.Request().Context()

	if resp, skipped := c.isProtectedContainer(ctx, containerID, "remove"); skipped {
		return resp, nil
	}

	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	opts := service.RemoveContainerOptions{
		ContainerID: containerID,
		Force:       true,
	}

	response, err := service.RemoveContainer(dockerService, c.logger, opts)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &response, nil
}
