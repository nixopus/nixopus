package controller

import (
	"net/http"

	"github.com/docker/docker/api/types/filters"
	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type PruneImagesRequest struct {
	Until    string `json:"until,omitempty"`
	Label    string `json:"label,omitempty"`
	Dangling bool   `json:"dangling,omitempty"`
}

func (c *ContainerController) PruneImages(f fuego.ContextWithBody[PruneImagesRequest]) (*shared_types.Response, error) {
	req, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}
	filterArgs := filters.NewArgs()
	if req.Until != "" {
		filterArgs.Add("until", req.Until)
	}
	if req.Label != "" {
		filterArgs.Add("label", req.Label)
	}
	if req.Dangling {
		filterArgs.Add("dangling", "true")
	}

	pruneReport, err := c.dockerService.PruneImages(filterArgs)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Images pruned successfully",
		Data:    pruneReport,
	}, nil
}
