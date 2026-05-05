package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/container/service"
	"github.com/nixopus/nixopus/api/internal/features/container/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

type PruneImagesRequest struct {
	Until    string `json:"until,omitempty"`
	Label    string `json:"label,omitempty"`
	Dangling bool   `json:"dangling,omitempty"`
}

func (c *ContainerController) PruneImages(f fuego.ContextWithBody[PruneImagesRequest]) (*types.PruneImagesResponse, error) {
	req, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("container: PruneImages invalid body: %v", err), "")
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	ctx := f.Request().Context()
	ctxStr := fmt.Sprintf("until=%q label=%q dangling=%v", req.Until, req.Label, req.Dangling)
	c.logger.Log(logger.Info, "container: PruneImages", ctxStr)

	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: PruneImages docker service: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	opts := service.PruneImagesOptions{
		Until:    req.Until,
		Label:    req.Label,
		Dangling: req.Dangling,
	}

	response, err := service.PruneImages(dockerService, c.logger, opts)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: PruneImages: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "container: PruneImages ok", ctxStr)
	return &response, nil
}
