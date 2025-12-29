package controller

import (
	"net/http"

	"github.com/docker/docker/api/types/filters"
	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
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

	// Convert Docker's DeleteResponse to our typed response
	imagesDeleted := make([]types.ImageDeleteResponse, len(pruneReport.ImagesDeleted))
	for i, img := range pruneReport.ImagesDeleted {
		imagesDeleted[i] = types.ImageDeleteResponse{
			Untagged: img.Untagged,
			Deleted:  img.Deleted,
		}
	}

	return &types.PruneImagesResponse{
		Status:  "success",
		Message: "Images pruned successfully",
		Data: types.PruneImagesResponseData{
			ImagesDeleted:  imagesDeleted,
			SpaceReclaimed: pruneReport.SpaceReclaimed,
		},
	}, nil
}
