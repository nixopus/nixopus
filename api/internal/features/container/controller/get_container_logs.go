package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/container/service"
	"github.com/nixopus/nixopus/api/internal/features/container/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *ContainerController) GetContainerLogs(f fuego.ContextWithBody[types.ContainerLogsRequest]) (*types.ContainerLogsResponse, error) {
	req, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("container: GetContainerLogs invalid body: %v", err), "")
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	if req.ID == "" {
		req.ID = f.PathParam("container_id")
	}
	if _, err := uuid.Parse(req.ID); err != nil {
		c.logger.Log(logger.Debug, "container: GetContainerLogs invalid id", req.ID)
		return nil, fuego.BadRequestError{Detail: "container_id must be a valid UUID"}
	}

	_, r := f.Response(), f.Request()
	ctx := r.Context()
	orgID := utils.GetOrganizationID(r)
	ctxStr := fmt.Sprintf("container_id=%s org_id=%s tail=%d follow=%v", req.ID, orgID, req.Tail, req.Follow)
	c.logger.Log(logger.Info, "container: GetContainerLogs", ctxStr)

	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: GetContainerLogs docker service: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	decodedLogs, err := service.GetContainerLogs(
		ctx,
		c.store,
		dockerService,
		c.logger,
		service.ContainerLogsOptions{
			ContainerID:    req.ID,
			OrganizationID: orgID.String(),
			Follow:         req.Follow,
			Tail:           req.Tail,
			Since:          req.Since,
			Until:          req.Until,
			Stdout:         req.Stdout,
			Stderr:         req.Stderr,
		},
	)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: GetContainerLogs: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "container: GetContainerLogs ok", fmt.Sprintf("%s bytes=%d", ctxStr, len(decodedLogs)))
	return &types.ContainerLogsResponse{
		Status:  "success",
		Message: "Container logs fetched successfully",
		Data:    decodedLogs,
	}, nil
}
