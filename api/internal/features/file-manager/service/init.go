package service

import (
	"context"
	"fmt"

	"github.com/melbahja/goph"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
)

type FileManagerService struct {
	logger logger.Logger
	Ctx    context.Context
	sshpkg *goph.Client
}

func NewFileManagerService(ctx context.Context, logger logger.Logger) *FileManagerService {
	client, err := ssh.NewSSH().ConnectWithPassword()
	if err != nil {
		fmt.Printf("Failed to create ssh client in file manager")
		return &FileManagerService{}
	}
	return &FileManagerService{
		logger: logger,
		Ctx:    ctx,
		sshpkg: client,
	}
}
