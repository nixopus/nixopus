package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/file-manager/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/uptrace/bun"
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
	db *bun.DB,
) *FileManagerController {
	return &FileManagerController{
		service:      service.NewFileManagerService(l, db),
		ctx:          ctx,
		logger:       l,
		notification: notificationManager,
	}
}
