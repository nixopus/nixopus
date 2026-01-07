package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *ContainerController) GetContainerLogs(f fuego.ContextWithBody[types.ContainerLogsRequest]) (*types.ContainerLogsResponse, error) {
	req, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	_, r := f.Response(), f.Request()
	orgID := utils.GetOrganizationID(r)

	decodedLogs, err := service.GetContainerLogs(
		c.ctx,
		c.store,
		c.dockerService,
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
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ContainerLogsResponse{
		Status:  "success",
		Message: "Container logs fetched successfully",
		Data:    decodedLogs,
	}, nil
}
