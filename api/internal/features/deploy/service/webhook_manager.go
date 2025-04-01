package service

import (
	"fmt"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *DeployService) HandleGithubWebhook(payload shared_types.WebhookPayload) error {
	parts := strings.Split(payload.Repository.FullName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository name format")
	}
	repositoryID := payload.Repository.ID

	branch := strings.TrimPrefix(payload.Ref, "refs/heads/")

	// check if we have an application for this repository and branch
	applications, err := s.storage.GetApplicationByRepositoryIDAndBranch(repositoryID, branch)
	if err != nil {
		return fmt.Errorf("failed to get application: %w", err)
	}

	if len(applications) == 0 {
		return fmt.Errorf("application not found")
	}

	// Process each application that matches the repository and branch
	for _, application := range applications {
		// Check if the branch is the same as the one in the application
		if application.Branch != branch {
			continue
		}

		// set the force flag to true to force the deployment
		deployment := &types.UpdateDeploymentRequest{
			ID:    application.ID,
			Force: true,
		}

		application = createApplicationFromExistingApplicationAndUpdateRequest(application, deployment)

		deploy_config, err := s.prepareDeploymentConfig(application, application.UserID, shared_types.DeploymentTypeUpdate, false, false)
		if err != nil {
			return err
		}

		go s.StartDeploymentInBackground(deploy_config)
		s.logger.Log(logger.Info, types.LogDeploymentStarted, "")
	}

	return nil
}
