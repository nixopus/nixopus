package service

import (
	"context"
	"fmt"

	"github.com/melbahja/goph"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type FileManagerService struct {
	logger logger.Logger
	store  *shared_storage.Store
}

// getSSHClient gets the SSH client from context (organization-specific)
func (f *FileManagerService) getSSHClient(ctx context.Context) (*goph.Client, error) {
	manager, err := ssh.GetSSHManagerFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH manager: %w", err)
	}
	client, err := manager.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to connect via SSH: %w", err)
	}
	return client, nil
}

func NewFileManagerService(ctx context.Context, store *shared_storage.Store, logger logger.Logger) *FileManagerService {
	return &FileManagerService{
		logger: logger,
		store:  store,
	}
}
