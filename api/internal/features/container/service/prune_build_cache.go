package service

import (
	"github.com/docker/docker/api/types"
	container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// PruneBuildCacheOptions contains options for pruning build cache
type PruneBuildCacheOptions struct {
	All bool
}

// PruneBuildCache prunes Docker build cache
func PruneBuildCache(
	dockerService *docker.DockerService,
	l logger.Logger,
	opts PruneBuildCacheOptions,
) (container_types.MessageResponse, error) {
	err := dockerService.PruneBuildCache(types.BuildCachePruneOptions{
		All: opts.All,
	})
	if err != nil {
		l.Log(logger.Error, err.Error(), "")
		return container_types.MessageResponse{}, err
	}

	return container_types.MessageResponse{
		Status:  "success",
		Message: "Build cache pruned successfully",
	}, nil
}
