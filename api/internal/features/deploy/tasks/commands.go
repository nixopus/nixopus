package tasks

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *TaskService) runCommands(ctx context.Context, applicationID uuid.UUID, deploymentConfigID uuid.UUID, command string, commandType string) error {
	taskCtx := s.NewTaskContext(shared_types.TaskPayload{
		Application: shared_types.Application{
			ID: applicationID,
		},
		ApplicationDeployment: shared_types.ApplicationDeployment{
			ID: deploymentConfigID,
		},
	})
	taskCtx.AddLog(fmt.Sprintf("Running %s commands %v", commandType, command))

	if command == "" {
		return nil
	}

	manager, err := ssh.GetSSHManagerFromContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to get SSH manager: %w", err)
	}
	output, err := manager.RunCommand(command)
	if err != nil {
		taskCtx.AddLog(fmt.Sprintf("Error while running %s command %v", commandType, output))
		return err
	}

	if output != "" {
		taskCtx.AddLog(fmt.Sprintf("%s command resulted in %v", commandType, output))
	}

	return nil
}

func (s *TaskService) PrerunCommands(ctx context.Context, d shared_types.TaskPayload) error {
	// Create context with organization ID from TaskPayload
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, d.Application.OrganizationID.String())
	return s.runCommands(orgCtx, d.Application.ID, d.ApplicationDeployment.ID,
		d.Application.PreRunCommand, "pre run")
}

func (s *TaskService) PostRunCommands(ctx context.Context, d shared_types.TaskPayload) error {
	// Create context with organization ID from TaskPayload
	orgCtx := context.WithValue(ctx, shared_types.OrganizationIDKey, d.Application.OrganizationID.String())
	return s.runCommands(orgCtx, d.Application.ID, d.ApplicationDeployment.ID,
		d.Application.PostRunCommand, "post run")
}
