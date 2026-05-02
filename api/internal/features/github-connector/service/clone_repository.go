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
	repo, err := s.GetRepositoryByID(c.UserID, c.RepoID)
	if err != nil {
		s.Logger.Log(logger.Error, fmt.Sprintf("Failed to get repository details: %s", err.Error()), "")
		return "", err
	}
	repoURL := repo.CloneURL

	s.Logger.Log(logger.Info, fmt.Sprintf("Cloning repository %s", repoURL), c.UserID)

	connectors, err := s.Storage.GetAllConnectors(c.UserID)
	if err != nil {
		s.Logger.Log(logger.Error, err.Error(), "")
		return "", err
	}

	if len(connectors) == 0 {
		s.Logger.Log(logger.Error, "No connectors found for user", c.UserID)
		return "", fmt.Errorf("no connectors found for user %s", c.UserID)
	}

	installationID := connectors[0].InstallationID
	jwt := gh.GenerateJwt(&connectors[0])

	accessToken, err := gh.InstallationToken(jwt, installationID)
	if err != nil {
		s.Logger.Log(logger.Error, fmt.Sprintf("Failed to get installation token: %s", err.Error()), "")
		return "", err
	}

	authenticatedURL, err := gh.CreateAuthenticatedRepoURL(repoURL, accessToken)
	if err != nil {
		s.Logger.Log(logger.Error, fmt.Sprintf("Failed to create authenticated URL: %s", err.Error()), "")
		return "", err
	}

	var latestCommitHash string
	if commitHash != nil {
		latestCommitHash = *commitHash
	} else {
		latestCommitHash, err = gh.LatestCommitHash(s.Logger, repoURL, accessToken)
		if err != nil {
			s.Logger.Log(logger.Error, fmt.Sprintf("Failed to get latest commit hash: %s", err.Error()), "")
			return "", err
		}
	}

	clonePath, shouldPull, err := s.GetClonePath(ctx, c.UserID, c.Environment, c.ApplicationID)
	if err != nil {
		s.Logger.Log(logger.Error, fmt.Sprintf("Failed to get clone path: %s", err.Error()), "")
		return "", err
	}

	s.Logger.Log(logger.Info, fmt.Sprintf("Clone path: %s", clonePath), "")

	gitClient, err := s.getGitClient(ctx)
	if err != nil {
		s.Logger.Log(logger.Error, fmt.Sprintf("Failed to get git client: %s", err.Error()), "")
		return "", err
	}

	if c.DeploymentType == shared_types.DeploymentTypeRollback {
		s.Logger.Log(logger.Info, "Rolling back repository", c.UserID)
		err = gitClient.SetHeadToCommitHash(authenticatedURL, clonePath, latestCommitHash)
		if err != nil {
			s.Logger.Log(logger.Error, fmt.Sprintf("Failed to rollback repository: %s", err.Error()), "")
			return "", err
		}
	} else {
		if !shouldPull {
			s.Logger.Log(logger.Info, "Cloning repository", c.UserID)
			err = gitClient.Clone(authenticatedURL, clonePath)
			if err != nil {
				s.Logger.Log(logger.Error, fmt.Sprintf("Failed to clone repository: %s", err.Error()), "")
				return "", err
			}
		} else {
			if err := git.HandlePullWithClient(s.Logger, gitClient, authenticatedURL, clonePath, c.UserID); err != nil {
				return "", err
			}
		}

		if c.Branch != "" {
			s.Logger.Log(logger.Info, fmt.Sprintf("Switching to branch %s", c.Branch), c.UserID)
			err = gitClient.SwitchBranch(clonePath, c.Branch)
			if err != nil {
				s.Logger.Log(logger.Error, fmt.Sprintf("Failed to switch to branch %s: %s", c.Branch, err.Error()), "")
				return "", err
			}
		}
	}

	s.Logger.Log(logger.Info, fmt.Sprintf("Context loaded successfully %s", repoURL), c.UserID)
	return clonePath, nil
}
