package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/file-manager/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type FileManagerController struct {
	service  *service.FileManagerService
	ctx      context.Context
	logger   logger.Logger
	notifier shared_types.Notifier
}

func NewFileManagerController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notifier shared_types.Notifier,
) *FileManagerController {
	return &FileManagerController{
		service:  service.NewFileManagerService(ctx, store, l),
		ctx:      ctx,
		logger:   l,
		notifier: notifier,
	}
}
