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
	ssh       *ssh.SSH
}

func NewGithubConnectorService(store *shared_storage.Store, ctx context.Context, l logger.Logger, GithubConnectorRepository storage.GithubConnectorRepository) *GithubConnectorService {
	sshService := ssh.NewSSH()
	return &GithubConnectorService{
		store:     store,
		ctx:       ctx,
		logger:    l,
		storage:   GithubConnectorRepository,
		gitClient: NewDefaultGitClient(l, sshService),
		ssh:       sshService,
	}
}
