package service

import (
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
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

// CloneRepository clones the specified repository for the given user and environment.
//
// The method takes in the repository ID, user ID, and environment as parameters.
// It first retrieves the repository URL from the given ID and user ID.
// Then, it retrieves the connectors associated with the user ID and uses the
// first connector to generate a JWT token. The token is then used to get an
// installation token, which is used to create an authenticated URL for the
// repository.
// Finally, the method clones the repository using the authenticated URL and
// returns the path to the cloned repository.
//
// If any errors occur during the process, the method logs the error and
// returns the error.
func (s *GithubConnectorService) CloneRepository(c CloneRepositoryConfig, commitHash *string) (string, error) {
	// Fetch repository directly by ID based on installation
	repo, err := s.GetGithubRepositoryByID(c.UserID, c.RepoID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to get repository details: %s", err.Error()), "")
		return "", err
	}
	repo_url := repo.CloneURL

	s.logger.Log(logger.Info, fmt.Sprintf("Cloning repository %s", repo_url), c.UserID)

	connectors, err := s.storage.GetAllConnectors(c.UserID)
	if err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return "", err
	}

	if len(connectors) == 0 {
		s.logger.Log(logger.Error, "No connectors found for user", c.UserID)
		return "", nil
	}

	// TODO: we will need to handle multiple connectors here
	installation_id := connectors[0].InstallationID
	jwt := GenerateJwt(&connectors[0])

	accessToken, err := s.getInstallationToken(jwt, installation_id)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to get installation token: %s", err.Error()), "")
		return "", err
	}

	authenticatedURL, err := s.CreateAuthenticatedRepoURL(repo_url, accessToken)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to create authenticated URL: %s", err.Error()), "")
		return "", err
	}
	var latestCommitHash string

	if commitHash != nil {
		latestCommitHash = *commitHash
	} else {
		latestCommitHash, err = s.gitClient.GetLatestCommitHash(repo_url, accessToken)
	}

	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to get latest commit hash: %s", err.Error()), "")
		return "", err
	}

	clonePath, should_pull, err := s.GetClonePath(c.UserID, c.Environment, c.ApplicationID)

	s.logger.Log(logger.Info, fmt.Sprintf("Clone path: %s", clonePath), "")

	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to get clone path: %s", err.Error()), "")
		return "", err
	}

	if c.DeploymentType == shared_types.DeploymentTypeRollback {
		s.logger.Log(logger.Info, "Rolling back repository", c.UserID)
		err = s.gitClient.SetHeadToCommitHash(authenticatedURL, clonePath, latestCommitHash)
		if err != nil {
			s.logger.Log(logger.Error, fmt.Sprintf("Failed to rollback repository: %s", err.Error()), "")
			return "", err
		}
	} else {
		if !should_pull {
			s.logger.Log(logger.Info, "Cloning repository", c.UserID)
			err = s.gitClient.Clone(authenticatedURL, clonePath)
			if err != nil {
				s.logger.Log(logger.Error, fmt.Sprintf("Failed to clone repository: %s", err.Error()), "")
				return "", err
			}
		} else {
			if err := s.handleGitPull(authenticatedURL, clonePath, c.UserID); err != nil {
				return "", err
			}
		}

		if c.Branch != "" {
			s.logger.Log(logger.Info, fmt.Sprintf("Switching to branch %s", c.Branch), c.UserID)
			err = s.gitClient.SwitchBranch(clonePath, c.Branch)
			if err != nil {
				s.logger.Log(logger.Error, fmt.Sprintf("Failed to switch to branch %s: %s", c.Branch, err.Error()), "")
				return "", err
			}
		}
	}

	s.logger.Log(logger.Info, fmt.Sprintf("Context loaded successfully %s", repo_url), c.UserID)
	return clonePath, nil
}
