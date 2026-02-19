package service

import (
	"context"
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type GithubConnectorService struct {
	store   *shared_storage.Store
	ctx     context.Context
	logger  logger.Logger
	storage storage.GithubConnectorRepository
}

// getSSHManager gets the SSH manager from context (organization-specific)
func (s *GithubConnectorService) getSSHManager(ctx context.Context) (*ssh.SSHManager, error) {
	return ssh.GetSSHManagerFromContext(ctx)
}

// getGitClient creates or returns a GitClient with the current SSH client from context
func (s *GithubConnectorService) getGitClient(ctx context.Context) (GitClient, error) {
	sshManager, err := s.getSSHManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH manager: %w", err)
	}
	sshClient, err := sshManager.GetDefaultSSH()
	if err != nil {
		return nil, fmt.Errorf("failed to get default SSH client: %w", err)
	}
	return NewDefaultGitClient(s.logger, sshClient), nil
}

func NewGithubConnectorService(store *shared_storage.Store, ctx context.Context, l logger.Logger, GithubConnectorRepository storage.GithubConnectorRepository) *GithubConnectorService {
	return &GithubConnectorService{
		store:   store,
		ctx:     ctx,
		logger:  l,
		storage: GithubConnectorRepository,
	}
}

func (s *GithubConnectorService) RemoveRepository(ctx context.Context, repoPath string) error {
	gitClient, err := s.getGitClient(ctx)
	if err != nil {
		return err
	}
	return gitClient.RemoveRepository(repoPath)
}
