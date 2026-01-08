package service

import (
	"github.com/docker/docker/api/types/filters"
	container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// PruneImagesOptions contains options for pruning images
type PruneImagesOptions struct {
	Until    string
	Label    string
	Dangling bool
}

// PruneImages prunes Docker images with optional filtering
func PruneImages(
	dockerService *docker.DockerService,
	l logger.Logger,
	opts PruneImagesOptions,
) (container_types.PruneImagesResponse, error) {
	filterArgs := filters.NewArgs()
	if opts.Until != "" {
		filterArgs.Add("until", opts.Until)
	}
	if opts.Label != "" {
		filterArgs.Add("label", opts.Label)
	}
	if opts.Dangling {
		filterArgs.Add("dangling", "true")
	}

	pruneReport, err := dockerService.PruneImages(filterArgs)
	if err != nil {
		l.Log(logger.Error, err.Error(), "")
		return container_types.PruneImagesResponse{}, err
	}

	// Convert Docker's DeleteResponse to our typed response
	imagesDeleted := make([]container_types.ImageDeleteResponse, len(pruneReport.ImagesDeleted))
	for i, img := range pruneReport.ImagesDeleted {
		imagesDeleted[i] = container_types.ImageDeleteResponse{
			Untagged: img.Untagged,
			Deleted:  img.Deleted,
		}
	}

	return container_types.PruneImagesResponse{
		Status:  "success",
		Message: "Images pruned successfully",
		Data: container_types.PruneImagesResponseData{
			ImagesDeleted:  imagesDeleted,
			SpaceReclaimed: pruneReport.SpaceReclaimed,
		},
	}, nil
}
