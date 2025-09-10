package tasks

import (
	"fmt"
	"strconv"

	github_service "github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type CloneConfig struct {
	shared_types.TaskPayload
	DeploymentType string
	TaskContext    *TaskContext
}

func (t *TaskService) Clone(cloneConfig CloneConfig) (string, error) {
	repoID, err := strconv.ParseInt(cloneConfig.Application.Repository, 10, 64)
	if err != nil {
		return "", fmt.Errorf("failed to parse repository id: %w", err)
	}
	cloneRepositoryConfig := github_service.CloneRepositoryConfig{
		RepoID:         uint64(repoID),
		UserID:         cloneConfig.Application.UserID.String(),
		Environment:    string(cloneConfig.Application.Environment),
		DeploymentID:   cloneConfig.ApplicationDeployment.ID.String(),
		DeploymentType: cloneConfig.DeploymentType,
		Branch:         cloneConfig.Application.Branch,
		ApplicationID:  cloneConfig.Application.ID.String(),
	}
	// we will pass the commit hash to the clone repository function for rollback feature otherwise it will clone the latest commit
	repoPath, err := t.Github_service.CloneRepository(cloneRepositoryConfig, &cloneConfig.ApplicationDeployment.CommitHash)
	if err != nil {
		return "", fmt.Errorf("failed to clone repository: %w", err)
	}
	return repoPath, nil
}
