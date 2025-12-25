package controller

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

type ContainerController struct {
	store         *shared_storage.Store
	dockerService *docker.DockerService
	ctx           context.Context
	logger        logger.Logger
	notification  *notification.NotificationManager
}

func NewContainerController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notificationManager *notification.NotificationManager,
) (*ContainerController, error) {
	dockerService, err := docker.GetDockerManager().GetDefaultService()
	if err != nil {
		return nil, fmt.Errorf("failed to get default docker service: %w", err)
	}
	if dockerService == nil {
		return nil, fmt.Errorf("docker service is nil")
	}
	return &ContainerController{
		store:         store,
		dockerService: dockerService,
		ctx:           ctx,
		logger:        l,
		notification:  notificationManager,
	}, nil
}

func (c *ContainerController) isProtectedContainer(containerID string, action string) (*types.ContainerActionResponse, bool) {
	details, err := c.dockerService.GetContainerById(containerID)
	if err != nil {
		return nil, false
	}
	name := strings.ToLower(details.Name)
	if strings.Contains(name, "nixopus") {
		c.logger.Log(logger.Info, fmt.Sprintf("Skipping %s for protected container", action), details.Name)
		return &types.ContainerActionResponse{
			Status:  "success",
			Message: "Operation skipped for protected container",
			Data:    types.ContainerStatusData{Status: "skipped"},
		}, true
	}
	return nil, false
}

// getOrganizationSettings retrieves organization settings with defaults
func (c *ContainerController) getOrganizationSettings(r *http.Request) shared_types.OrganizationSettingsData {
	orgID := utils.GetOrganizationID(r)
	if orgID == uuid.Nil {
		return shared_types.DefaultOrganizationSettingsData()
	}

	orgStore := storage.OrganizationStore{DB: c.store.DB, Ctx: c.ctx}
	settings, err := orgStore.GetOrganizationSettings(orgID.String())
	if err != nil || settings == nil {
		return shared_types.DefaultOrganizationSettingsData()
	}

	// Merge with defaults to ensure all fields are set
	defaults := shared_types.DefaultOrganizationSettingsData()
	result := shared_types.OrganizationSettingsData{
		WebsocketReconnectAttempts:       settings.Settings.WebsocketReconnectAttempts,
		WebsocketReconnectInterval:       settings.Settings.WebsocketReconnectInterval,
		ApiRetryAttempts:                 settings.Settings.ApiRetryAttempts,
		DisableApiCache:                  settings.Settings.DisableApiCache,
		ContainerLogTailLines:            defaults.ContainerLogTailLines,
		ContainerDefaultRestartPolicy:    defaults.ContainerDefaultRestartPolicy,
		ContainerStopTimeout:             defaults.ContainerStopTimeout,
		ContainerAutoPruneDanglingImages: defaults.ContainerAutoPruneDanglingImages,
		ContainerAutoPruneBuildCache:     defaults.ContainerAutoPruneBuildCache,
	}

	if settings.Settings.ContainerLogTailLines != nil {
		result.ContainerLogTailLines = settings.Settings.ContainerLogTailLines
	}
	if settings.Settings.ContainerDefaultRestartPolicy != nil {
		result.ContainerDefaultRestartPolicy = settings.Settings.ContainerDefaultRestartPolicy
	}
	if settings.Settings.ContainerStopTimeout != nil {
		result.ContainerStopTimeout = settings.Settings.ContainerStopTimeout
	}
	if settings.Settings.ContainerAutoPruneDanglingImages != nil {
		result.ContainerAutoPruneDanglingImages = settings.Settings.ContainerAutoPruneDanglingImages
	}
	if settings.Settings.ContainerAutoPruneBuildCache != nil {
		result.ContainerAutoPruneBuildCache = settings.Settings.ContainerAutoPruneBuildCache
	}

	return result
}
