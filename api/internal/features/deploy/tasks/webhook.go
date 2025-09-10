package tasks

import (
	"fmt"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"strings"
)

func (t *TaskService) EnqueueWebhookTask(payload shared_types.WebhookPayload) error {
	parts := strings.Split(payload.Repository.FullName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository name format")
	}
	repositoryID := payload.Repository.ID

	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	applications, err := t.Storage.GetApplicationByRepositoryIDAndBranch(repositoryID, branch)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}

	if len(applications) == 0 {
		return fmt.Errorf("application not found")
	}

	for _, application := range applications {
		if application.Branch != branch {
			continue
		}

		deployment := &types.UpdateDeploymentRequest{
			ID:                   application.ID,
			Force:                true,
			PreRunCommand:        application.PreRunCommand,
			PostRunCommand:       application.PostRunCommand,
			BuildVariables:       GetMapFromString(application.BuildVariables),
			EnvironmentVariables: GetMapFromString(application.EnvironmentVariables),
			Port:                 application.Port,
			DockerfilePath:       application.DockerfilePath,
			BasePath:             application.BasePath,
		}

		_, err := t.UpdateDeployment(deployment, application.UserID, application.OrganizationID)
		if err != nil {
			t.Logger.Log(logger.Error, "failed to update deployment for webhook", err.Error())
			continue
		}

		t.Logger.Log(logger.Info, types.LogDeploymentStarted, "")
	}

	return nil
}
