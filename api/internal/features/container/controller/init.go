package controller

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/container/types"
	"github.com/nixopus/nixopus/api/internal/features/deploy/docker"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

type ContainerController struct {
	store    *shared_storage.Store
	ctx      context.Context
	logger   logger.Logger
	notifier shared_types.Notifier
}

func NewContainerController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notifier shared_types.Notifier,
) (*ContainerController, error) {
	return &ContainerController{
		store:    store,
		ctx:      ctx,
		logger:   l,
		notifier: notifier,
	}, nil
}

func (c *ContainerController) getDockerService(ctx context.Context) (docker.DockerRepository, error) {
	dockerService, err := docker.GetDockerServiceFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get docker service: %w", err)
	}
	if dockerService == nil {
		return nil, fmt.Errorf("docker service is nil")
	}
	return dockerService, nil
}

func (c *ContainerController) isProtectedContainer(ctx context.Context, containerID string, action string) (*types.ContainerActionResponse, bool) {
	dockerService, err := c.getDockerService(ctx)
	if err != nil {
		return nil, false
	}
	details, err := dockerService.GetContainerById(containerID)
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
