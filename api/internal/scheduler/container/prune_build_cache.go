package container

import (
	"context"
	"fmt"

	docker_types "github.com/docker/docker/api/types"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	PruneBuildCacheJobName = "prune_build_cache"
)

// PruneBuildCacheJob handles cleanup of Docker build cache
type PruneBuildCacheJob struct {
	logger logger.Logger
}

// NewPruneBuildCacheJob creates a new prune build cache job
func NewPruneBuildCacheJob(l logger.Logger) *PruneBuildCacheJob {
	return &PruneBuildCacheJob{
		logger: l,
	}
}

// Name returns the job name
func (j *PruneBuildCacheJob) Name() string {
	return PruneBuildCacheJobName
}

// IsEnabled checks if auto prune build cache is enabled for this organization
func (j *PruneBuildCacheJob) IsEnabled(settings *types.OrganizationSettingsData) bool {
	return IsEnabledOrDefault(settings.ContainerAutoPruneBuildCache)
}

// GetRetentionDays returns 0 as this job doesn't use retention days
func (j *PruneBuildCacheJob) GetRetentionDays(settings *types.OrganizationSettingsData) int {
	return 0
}

// Run executes the prune build cache job
// Note: Build cache is system wide, so this runs once when any org has it enabled
func (j *PruneBuildCacheJob) Run(ctx context.Context, orgID uuid.UUID) error {
	j.logger.Log(
		logger.Info,
		fmt.Sprintf("Running build cache prune (triggered by org %s)", orgID),
		"",
	)

	// Get Docker service
	dockerService, err := docker.GetDockerManager().GetDefaultService()
	if err != nil {
		return fmt.Errorf("failed to get docker service: %w", err)
	}
	if dockerService == nil {
		return fmt.Errorf("docker service is nil")
	}

	// Prune all build cache
	err = dockerService.PruneBuildCache(docker_types.BuildCachePruneOptions{
		All: true,
	})
	if err != nil {
		return fmt.Errorf("failed to prune build cache: %w", err)
	}

	j.logger.Log(
		logger.Info,
		fmt.Sprintf("Build cache pruned successfully (triggered by org %s)", orgID),
		"",
	)

	return nil
}
