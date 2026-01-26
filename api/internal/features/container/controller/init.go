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
	orgID, err := utils.GetOrCreateOrganizationID(r.Context(), r, &shared_storage.App{Store: c.store, Ctx: c.ctx})
	if err != nil || orgID == uuid.Nil {
		return shared_types.DefaultOrganizationSettingsData()
	}

	settings, err := utils.GetOrganizationSettings(c.ctx, c.store.DB, orgID)
	if err != nil {
		return shared_types.DefaultOrganizationSettingsData()
	}

	return settings
}
