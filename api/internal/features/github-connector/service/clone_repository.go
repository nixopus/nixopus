package service

import (
	"context"
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/github-connector/service/git"
	gh "github.com/nixopus/nixopus/api/internal/features/github-connector/service/github"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

type CloneRepositoryConfig struct {
	RepoID         uint64
	UserID         string
	Environment    string
	DeploymentID   string
	DeploymentType string
	Branch         string
	ApplicationID  string
}

func (s *GithubConnectorService) CloneRepository(ctx context.Context, c CloneRepositoryConfig, commitHash *string) (string, error) {
	ctxStr := fmt.Sprintf("user_id=%s repo_id=%d env=%s app_id=%s", c.UserID, c.RepoID, c.Environment, c.ApplicationID)
	repo, err := s.GetRepositoryByID(c.UserID, c.RepoID)
	if err != nil {
		s.Logger.Log(logger.Error, fmt.Sprintf("github connector service: CloneRepository get repo: %v", err), ctxStr)
		return "", err
	}
	repoURL := repo.CloneURL

	s.Logger.Log(logger.Info, fmt.Sprintf("github connector service: CloneRepository start repo_url=%s", repoURL), ctxStr)

	connectors, err := s.Storage.GetAllConnectors(c.UserID)
	if err != nil {
		s.Logger.Log(logger.Error, fmt.Sprintf("github connector service: CloneRepository list connectors: %v", err), ctxStr)
		return "", err
	}

	if len(connectors) == 0 {
		s.Logger.Log(logger.Error, "github connector service: CloneRepository no connectors", ctxStr)
		return "", fmt.Errorf("no connectors found for user %s", c.UserID)
	}

	installationID := connectors[0].InstallationID
	jwt := gh.GenerateJwt(&connectors[0])

	accessToken, err := gh.InstallationToken(jwt, installationID)
	if err != nil {
		s.Logger.Log(logger.Error, fmt.Sprintf("github connector service: CloneRepository installation token: %v", err), ctxStr)
		return "", err
	}

	authenticatedURL, err := gh.CreateAuthenticatedRepoURL(repoURL, accessToken)
	if err != nil {
		s.Logger.Log(logger.Error, fmt.Sprintf("github connector service: CloneRepository authenticated URL: %v", err), ctxStr)
		return "", err
	}

	var latestCommitHash string
	if commitHash != nil {
		latestCommitHash = *commitHash
	} else {
		latestCommitHash, err = gh.LatestCommitHash(s.Logger, repoURL, accessToken)
		if err != nil {
			s.Logger.Log(logger.Error, fmt.Sprintf("github connector service: CloneRepository latest commit: %v", err), ctxStr)
			return "", err
		}
	}

	clonePath, shouldPull, err := s.GetClonePath(ctx, c.UserID, c.Environment, c.ApplicationID)
	if err != nil {
		s.Logger.Log(logger.Error, fmt.Sprintf("github connector service: CloneRepository clone path: %v", err), ctxStr)
		return "", err
	}

	s.Logger.Log(logger.Info, fmt.Sprintf("github connector service: CloneRepository clone_path=%s should_pull=%t", clonePath, shouldPull), ctxStr)

	gitClient, err := s.getGitClient(ctx)
	if err != nil {
		s.Logger.Log(logger.Error, fmt.Sprintf("github connector service: CloneRepository git client: %v", err), ctxStr)
		return "", err
	}

	if c.DeploymentType == shared_types.DeploymentTypeRollback {
		s.Logger.Log(logger.Info, "github connector service: CloneRepository rollback", ctxStr)
		err = gitClient.SetHeadToCommitHash(authenticatedURL, clonePath, latestCommitHash)
		if err != nil {
			s.Logger.Log(logger.Error, fmt.Sprintf("github connector service: CloneRepository rollback: %v", err), ctxStr)
			return "", err
		}
	} else {
		if !shouldPull {
			s.Logger.Log(logger.Info, "github connector service: CloneRepository clone", ctxStr)
			err = gitClient.Clone(authenticatedURL, clonePath)
			if err != nil {
				s.Logger.Log(logger.Error, fmt.Sprintf("github connector service: CloneRepository clone failed: %v", err), ctxStr)
				return "", err
			}
		} else {
			if err := git.HandlePullWithClient(s.Logger, gitClient, authenticatedURL, clonePath, c.UserID); err != nil {
				return "", err
			}
		}

		if c.Branch != "" {
			s.Logger.Log(logger.Info, fmt.Sprintf("github connector service: CloneRepository checkout branch=%s", c.Branch), ctxStr)
			err = gitClient.SwitchBranch(clonePath, c.Branch)
			if err != nil {
				s.Logger.Log(logger.Error, fmt.Sprintf("github connector service: CloneRepository branch %s: %v", c.Branch, err), ctxStr)
				return "", err
			}
		}
	}

	s.Logger.Log(logger.Info, "github connector service: CloneRepository ok", ctxStr)
	return clonePath, nil
}
