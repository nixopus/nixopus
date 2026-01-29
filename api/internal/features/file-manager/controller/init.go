package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/file-manager/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type FileManagerController struct {
	service      *service.FileManagerService
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
}

func NewFileManagerController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notificationManager *notification.NotificationManager,
) *FileManagerController {
	return &FileManagerController{
		service:      service.NewFileManagerService(ctx, store, l),
		ctx:          ctx,
		logger:       l,
		notification: notificationManager,
	}
}
