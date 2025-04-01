package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/file-manager/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
)

type FileManagerController struct {
	service      *service.FileManagerService
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
}

func NewFileManagerController(
	ctx context.Context,
	l logger.Logger,
	notificationManager *notification.NotificationManager,
) *FileManagerController {
	return &FileManagerController{
		service:      service.NewFileManagerService(ctx, l),
		ctx:          ctx,
		logger:       l,
		notification: notificationManager,
	}
}
