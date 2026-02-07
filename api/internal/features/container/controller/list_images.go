package controller

import (
	"net/http"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/go-fuego/fuego"
	container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"
)

type ListImagesRequest struct {
	All         bool   `json:"all,omitempty"`
	ContainerID string `json:"container_id,omitempty"`
	ImagePrefix string `json:"image_prefix,omitempty"`
}

func (c *ContainerController) ListImages(f fuego.ContextWithBody[ListImagesRequest]) (*container_types.ListImagesResponse, error) {
	req, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	ctx := f.Request().Context()
	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	filterArgs := filters.NewArgs()
	if req.ContainerID != "" {
		_, err := dockerService.GetContainerById(req.ContainerID)
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusNotFound,
			}
		}
	}

	if req.ImagePrefix != "" {
		pattern := req.ImagePrefix
		if !strings.HasSuffix(pattern, "*") {
			pattern += "*"
		}
		filterArgs.Add("reference", pattern)
	}

	images := dockerService.ListAllImages(image.ListOptions{
		All:     req.All,
		Filters: filterArgs,
	})

	if len(images) == 0 {
		return &container_types.ListImagesResponse{
			Status:  "success",
			Message: "No images found",
			Data:    []container_types.Image{},
		}, nil
	}

	var result []container_types.Image
	for _, img := range images {
		imageData := container_types.Image{
			ID:          img.ID,
			RepoTags:    img.RepoTags,
			RepoDigests: img.RepoDigests,
			Created:     img.Created,
			Size:        img.Size,
			SharedSize:  img.SharedSize,
			VirtualSize: img.VirtualSize,
			Labels:      img.Labels,
		}

		result = append(result, imageData)
	}

	return &container_types.ListImagesResponse{
		Status:  "success",
		Message: "Images listed successfully",
		Data:    result,
	}, nil
}
