package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/go-fuego/fuego"
	container_types "github.com/nixopus/nixopus/api/internal/features/container/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

type ListImagesRequest struct {
	All         bool   `json:"all,omitempty"`
	ContainerID string `json:"container_id,omitempty"`
	ImagePrefix string `json:"image_prefix,omitempty"`
}

func (c *ContainerController) ListImages(f fuego.ContextWithBody[ListImagesRequest]) (*container_types.ListImagesResponse, error) {
	req, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("container: ListImages invalid body: %v", err), "")
		return nil, fuego.BadRequestError{
			Detail: err.Error(),
			Err:    err,
		}
	}

	ctx := f.Request().Context()
	ctxStr := fmt.Sprintf("all=%v container_id=%s image_prefix=%q", req.All, req.ContainerID, req.ImagePrefix)
	c.logger.Log(logger.Info, "container: ListImages", ctxStr)

	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("container: ListImages docker service: %v", err), ctxStr)
		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	filterArgs := filters.NewArgs()
	if req.ContainerID != "" {
		_, err := dockerService.GetContainerById(req.ContainerID)
		if err != nil {
			c.logger.Log(logger.Debug, fmt.Sprintf("container: ListImages container not found: %v", err), ctxStr)
			return nil, fuego.NotFoundError{
				Detail: err.Error(),
				Err:    err,
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
		c.logger.Log(logger.Info, "container: ListImages ok", fmt.Sprintf("%s count=0", ctxStr))
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

	c.logger.Log(logger.Info, "container: ListImages ok", fmt.Sprintf("%s count=%d", ctxStr, len(result)))
	return &container_types.ListImagesResponse{
		Status:  "success",
		Message: "Images listed successfully",
		Data:    result,
	}, nil
}
