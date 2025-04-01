package service

import (
	"fmt"
)

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
