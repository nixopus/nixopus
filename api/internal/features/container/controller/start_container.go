package controller

import (
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (c *ContainerController) StartContainer(f fuego.ContextNoBody) (*types.ContainerActionResponse, error) {
	containerID := f.PathParam("container_id")

	if resp, skipped := c.isProtectedContainer(containerID, "start"); skipped {
		return resp, nil
	}

	err := c.dockerService.StartContainer(containerID, container.StartOptions{})
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ContainerActionResponse{
		Status:  "success",
		Message: "Container started successfully",
		Data:    types.ContainerStatusData{Status: "started"},
	}, nil
}
