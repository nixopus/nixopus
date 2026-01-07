package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (c *ContainerController) GetContainer(f fuego.ContextNoBody) (*types.GetContainerResponse, error) {
	containerID := f.PathParam("container_id")

	containerData, err := service.GetContainer(c.dockerService, c.logger, containerID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.GetContainerResponse{
		Status:  "success",
		Message: "Container fetched successfully",
		Data:    containerData,
	}, nil
}
