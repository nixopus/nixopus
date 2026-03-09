package service

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
)

type FileManagerService struct {
	logger logger.Logger
	store  *shared_storage.Store
}

func NewFileManagerService(ctx context.Context, store *shared_storage.Store, logger logger.Logger) *FileManagerService {
	return &FileManagerService{
		logger: logger,
		store:  store,
	}
}
