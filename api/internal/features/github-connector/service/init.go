package service

import (
	"context"
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/github-connector/service/git"
	gh "github.com/nixopus/nixopus/api/internal/features/github-connector/service/github"
	"github.com/nixopus/nixopus/api/internal/features/github-connector/storage"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
)

type GithubConnectorService struct {
	*gh.API
	store *shared_storage.Store
	ctx   context.Context

	gitClientProvider func(context.Context) (git.Git, error)
	clonePathProvider func(context.Context, string, string, string) (string, bool, error)
}

func NewGithubConnectorService(store *shared_storage.Store, ctx context.Context, l logger.Logger, githubConnectorRepository storage.GithubConnectorRepository) *GithubConnectorService {
	return &GithubConnectorService{
		API:   gh.NewAPI(githubConnectorRepository, l),
		store: store,
		ctx:   ctx,
	}
}

func (s *GithubConnectorService) getSSHManager(ctx context.Context) (*ssh.SSHManager, error) {
	return ssh.GetSSHManagerFromContext(ctx)
}

func (s *GithubConnectorService) getGitClient(ctx context.Context) (git.Git, error) {
	if s.gitClientProvider != nil {
		return s.gitClientProvider(ctx)
	}
	sshManager, err := s.getSSHManager(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH manager: %w", err)
	}
	return git.NewGit(s.Logger, sshManager), nil
}

func (s *GithubConnectorService) RemoveRepository(ctx context.Context, repoPath string) error {
	gitClient, err := s.getGitClient(ctx)
	if err != nil {
		return err
	}
	return gitClient.RemoveRepository(repoPath)
}

func (s *GithubConnectorService) CreateAuthenticatedRepoURL(repoURL, accessToken string) (string, error) {
	return gh.CreateAuthenticatedRepoURL(repoURL, accessToken)
}

func (s *GithubConnectorService) GetClonePath(ctx context.Context, userID, environment, applicationID string) (string, bool, error) {
	if s.clonePathProvider != nil {
		return s.clonePathProvider(ctx, userID, environment, applicationID)
	}
	return git.ResolveClonePath(ctx, s.Logger, userID, environment, applicationID)
}
