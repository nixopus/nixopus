package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/container/service"
	"github.com/nixopus/nixopus/api/internal/features/container/types"
)

type PruneImagesRequest struct {
	Until    string `json:"until,omitempty"`
	Label    string `json:"label,omitempty"`
	Dangling bool   `json:"dangling,omitempty"`
}

func (c *ContainerController) PruneImages(f fuego.ContextWithBody[PruneImagesRequest]) (*types.PruneImagesResponse, error) {
	req, err := f.Body()
	if err != nil {
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	ctx := f.Request().Context()
	dockerService, err := c.getDockerService(ctx)
	if err != nil {
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
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	return &response, nil
}
