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

func (c *ContainerController) GetContainer(f fuego.ContextNoBody) (*types.GetContainerResponse, error) {
	containerID := f.PathParam("container_id")
	if _, err := uuid.Parse(containerID); err != nil {
		c.logger.Log(logger.Debug, "container: GetContainer invalid id", containerID)
		return nil, fuego.BadRequestError{Detail: "container_id must be a valid UUID"}
	}
	ctx := f.Request().Context()
	ctxStr := fmt.Sprintf("container_id=%s", containerID)
	c.logger.Log(logger.Info, "container: GetContainer", ctxStr)

	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: GetContainer docker service: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	containerData, err := service.GetContainer(dockerService, c.logger, containerID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: GetContainer: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "container: GetContainer ok", ctxStr)
	return &types.GetContainerResponse{
		Status:  "success",
		Message: "Container fetched successfully",
		Data:    containerData,
	}, nil
}
