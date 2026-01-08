package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/service"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
)

type PruneImagesRequest struct {
	Until    string `json:"until,omitempty"`
	Label    string `json:"label,omitempty"`
	Dangling bool   `json:"dangling,omitempty"`
}

func (c *ContainerController) PruneImages(f fuego.ContextWithBody[PruneImagesRequest]) (*types.PruneImagesResponse, error) {
	req, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	opts := service.PruneImagesOptions{
		Until:    req.Until,
		Label:    req.Label,
		Dangling: req.Dangling,
	}

	response, err := service.PruneImages(c.dockerService, c.logger, opts)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &response, nil
}
