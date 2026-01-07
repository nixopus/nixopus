package service

import (
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// ListImagesOptions contains options for listing images
type ListImagesOptions struct {
	All         bool
	ContainerID string
	ImagePrefix string
}

// ListImages retrieves a list of Docker images with optional filtering
func ListImages(
	dockerService *docker.DockerService,
	l logger.Logger,
	opts ListImagesOptions,
) (container_types.ListImagesResponse, error) {
	filterArgs := filters.NewArgs()

	// Validate container ID if provided
	if opts.ContainerID != "" {
		_, err := dockerService.GetContainerById(opts.ContainerID)
		if err != nil {
			l.Log(logger.Error, err.Error(), "")
			return container_types.ListImagesResponse{}, err
		}
	}

	// Add image prefix filter if provided
	if opts.ImagePrefix != "" {
		pattern := opts.ImagePrefix
		if !strings.HasSuffix(pattern, "*") {
			pattern += "*"
		}
		filterArgs.Add("reference", pattern)
	}

	// List images from Docker
	images := dockerService.ListAllImages(image.ListOptions{
		All:     opts.All,
		Filters: filterArgs,
	})

	if len(images) == 0 {
		return container_types.ListImagesResponse{
			Status:  "success",
			Message: "No images found",
			Data:    []container_types.Image{},
		}, nil
	}

	// Transform Docker image summaries to our Image type
	var result []container_types.Image
	for _, img := range images {
		labels := img.Labels
		if labels == nil {
			labels = make(map[string]string)
		}
		repoTags := img.RepoTags
		if repoTags == nil {
			repoTags = []string{}
		}
		repoDigests := img.RepoDigests
		if repoDigests == nil {
			repoDigests = []string{}
		}

		imageData := container_types.Image{
			ID:          img.ID,
			RepoTags:    repoTags,
			RepoDigests: repoDigests,
			Created:     img.Created,
			Size:        img.Size,
			SharedSize:  img.SharedSize,
			VirtualSize: img.VirtualSize,
			Labels:      labels,
		}

		result = append(result, imageData)
	}

	return container_types.ListImagesResponse{
		Status:  "success",
		Message: "Images listed successfully",
		Data:    result,
	}, nil
}
