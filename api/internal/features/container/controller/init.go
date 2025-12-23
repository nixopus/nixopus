package controller

import (
	"context"
	"fmt"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
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
) *ContainerController {
	return &ContainerController{
		store:         store,
		dockerService: docker.NewDockerService(),
		ctx:           ctx,
		logger:        l,
		notification:  notificationManager,
	}
}

func (c *ContainerController) isProtectedContainer(containerID string, action string) (*shared_types.Response, bool) {
	details, err := c.dockerService.GetContainerById(containerID)
	if err != nil {
		return nil, false
	}
	name := strings.ToLower(details.Name)
	if strings.Contains(name, "nixopus") {
		c.logger.Log(logger.Info, fmt.Sprintf("Skipping %s for protected container", action), details.Name)
		return &shared_types.Response{
			Status:  "success",
			Message: "Operation skipped for protected container",
			Data:    map[string]string{"status": "skipped"},
		}, true
	}
	return nil, false
}
