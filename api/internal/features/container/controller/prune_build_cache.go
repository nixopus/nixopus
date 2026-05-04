package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/container/service"
	container_types "github.com/nixopus/nixopus/api/internal/features/container/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

type PruneBuildCacheRequest struct {
	All     bool   `json:"all,omitempty"`
	Filters string `json:"filters,omitempty"`
}

func (c *ContainerController) PruneBuildCache(f fuego.ContextWithBody[PruneBuildCacheRequest]) (*container_types.MessageResponse, error) {
	req, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("container: PruneBuildCache invalid body: %v", err), "")
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	ctx := f.Request().Context()
	ctxStr := fmt.Sprintf("all=%v filters=%q", req.All, req.Filters)
	c.logger.Log(logger.Info, "container: PruneBuildCache", ctxStr)

	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: PruneBuildCache docker service: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	opts := service.PruneBuildCacheOptions{
		All: req.All,
	}

	response, err := service.PruneBuildCache(dockerService, c.logger, opts)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: PruneBuildCache: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "container: PruneBuildCache ok", ctxStr)
	return &response, nil
}
