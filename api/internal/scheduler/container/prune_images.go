package container

import (
	"context"
	"fmt"

	"github.com/docker/docker/api/types/filters"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

const (
	PruneImagesJobName = "prune_dangling_images"
)

// PruneImagesJob handles cleanup of dangling Docker images
type PruneImagesJob struct {
	logger logger.Logger
}

// NewPruneImagesJob creates a new prune images job
func NewPruneImagesJob(l logger.Logger) *PruneImagesJob {
	return &PruneImagesJob{
		logger: l,
	}
}

// Name returns the job name
func (j *PruneImagesJob) Name() string {
	return PruneImagesJobName
}

// IsEnabled checks if auto prune is enabled for this organization
func (j *PruneImagesJob) IsEnabled(settings *types.OrganizationSettingsData) bool {
	return IsEnabledOrDefault(settings.ContainerAutoPruneDanglingImages)
}

// GetRetentionDays returns 0 as this job doesn't use retention days
func (j *PruneImagesJob) GetRetentionDays(settings *types.OrganizationSettingsData) int {
	return 0
}

// Run executes the prune images job
// Note: Docker images are system-wide, so this runs once when any org has it enabled
func (j *PruneImagesJob) Run(ctx context.Context, orgID uuid.UUID) error {
	j.logger.Log(
		logger.Info,
		fmt.Sprintf("Running dangling images prune (triggered by org %s)", orgID),
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

	// Prune dangling images only
	filterArgs := filters.NewArgs()
	filterArgs.Add("dangling", "true")

	pruneReport, err := dockerService.PruneImages(filterArgs)
	if err != nil {
		return fmt.Errorf("failed to prune images: %w", err)
	}

	j.logger.Log(
		logger.Info,
		fmt.Sprintf("Pruned %d dangling images, reclaimed %d bytes (triggered by org %s)",
			len(pruneReport.ImagesDeleted), pruneReport.SpaceReclaimed, orgID),
		"",
	)

	return nil
}
