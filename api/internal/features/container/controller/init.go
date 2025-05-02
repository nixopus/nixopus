package controller

import (
	"context"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
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
