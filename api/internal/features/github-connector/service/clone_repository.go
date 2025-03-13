package service

import (
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

type CloneRepositoryConfig struct {
	RepoID         uint64
	UserID         string
	Environment    string
	DeploymentID   string
	DeploymentType string
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
func (s *GithubConnectorService) CloneRepository(c CloneRepositoryConfig) (string, error) {
	_, repo_url, err := s.GetRepositoryDetailsFromId(c.RepoID, c.UserID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to get repository details: %s", err.Error()), "")
		return "", err
	}

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

	installation_id := connectors[0].InstallationID

	jwt := GenerateJwt(&connectors[0])

	accessToken, err := s.getInstallationToken(jwt, installation_id)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to get installation token: %s", err.Error()), "")
		return "", err
	}

	authenticatedURL, err := s.createAuthenticatedRepoURL(repo_url, accessToken)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to create authenticated URL: %s", err.Error()), "")
		return "", err
	}

	clonePath := s.getClonePath(c.UserID, c.Environment, repo_url, c.DeploymentID)

	// If deployment type is not provided, default to "create"
	if c.DeploymentType == "" {
		c.DeploymentType = "create"
	}

	err = s.gitClient.Clone(authenticatedURL, clonePath)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to clone repository: %s", err.Error()), "")
		return "", err
	}

	s.logger.Log(logger.Info, fmt.Sprintf("Successfully cloned repository %s", repo_url), c.UserID)
	return clonePath, nil
}

// GetRepositoryDetailsFromId retrieves the name and clone URL of a repository
// given its ID.
//
// Parameters:
//
//	repoID - the ID of the repository to retrieve.
//	UserID - the ID of the user whose repositories to search.
//
// Returns:
//
//	string - the name of the repository if found, otherwise an empty string.
//	string - the clone URL of the repository if found, otherwise an empty string.
//	error - an error if the repository is not found or if the method fails.
func (s *GithubConnectorService) GetRepositoryDetailsFromId(repoID uint64, UserID string) (string, string, error) {
	repositories, err := s.GetGithubRepositories(UserID)

	if err != nil {
		return "", "", err
	}

	for _, repository := range repositories {
		if repository.ID == repoID {
			return repository.Name, repository.CloneURL, nil
		}
	}

	return "", "", fmt.Errorf("repository not found")
}
