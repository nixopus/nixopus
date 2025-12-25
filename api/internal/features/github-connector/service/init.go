package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type GithubConnectorService struct {
	store     *shared_storage.Store
	ctx       context.Context
	logger    logger.Logger
	storage   storage.GithubConnectorRepository
	gitClient GitClient
	ssh       *ssh.SSHManager
}

func NewGithubConnectorService(store *shared_storage.Store, ctx context.Context, l logger.Logger, GithubConnectorRepository storage.GithubConnectorRepository) *GithubConnectorService {
	sshManager := ssh.GetSSHManager()
	sshClient, _ := sshManager.GetDefaultSSH() // Get default SSH client for GitClient
	return &GithubConnectorService{
		store:     store,
		ctx:       ctx,
		logger:    l,
		storage:   GithubConnectorRepository,
		gitClient: NewDefaultGitClient(l, sshClient),
		ssh:       sshManager,
	}
}

func (s *GithubConnectorService) RemoveRepository(repoPath string) error {
	return s.gitClient.RemoveRepository(repoPath)
}
