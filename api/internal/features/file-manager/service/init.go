package service

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/melbahja/goph"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"github.com/uptrace/bun"
)

type FileManagerService struct {
	logger logger.Logger
	db     *bun.DB
}

func NewFileManagerService(logger logger.Logger, db *bun.DB) *FileManagerService {
	return &FileManagerService{
		logger: logger,
		db:     db,
	}
}

// getSSHClient creates an SSH client using the active server from the database
// It extracts the organization ID from the request context (set by supertokens middleware)
func (f *FileManagerService) getSSHClient(ctx context.Context) (*goph.Client, error) {
	organizationID := utils.GetOrganizationIDFromContext(ctx)
	if organizationID == uuid.Nil {
		log.Printf("No organization ID in context, falling back to default SSH config")
		client, err := ssh.NewSSH().Connect()
		if err != nil {
			return nil, fmt.Errorf("failed to create default SSH client: %w", err)
		}
		return client, nil
	}

	sshConfig := ssh.NewSSHWithServer(f.db, ctx, organizationID)
	client, err := sshConfig.Connect()
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client with server config: %w", err)
	}

	return client, nil
}
