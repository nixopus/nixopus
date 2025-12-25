package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
)

func (s *DeployService) runCommands(applicationID uuid.UUID, deploymentConfigID uuid.UUID,
	command string, commandType string) error {
	s.addLog(applicationID, fmt.Sprintf("Running %s commands %v", commandType, command), deploymentConfigID)

	if command == "" {
		return nil
	}

	manager := ssh.GetSSHManager()
	output, err := manager.RunCommand(command)
	if err != nil {
		s.addLog(applicationID, fmt.Sprintf("Error while running %s command %v", commandType, output), deploymentConfigID)
		return err
	}

	if output != "" {
		s.addLog(applicationID, fmt.Sprintf("%s command resulted in %v", commandType, output), deploymentConfigID)
	}

	return nil
}

func (s *DeployService) PrerunCommands(d DeployerConfig) error {
	return s.runCommands(d.application.ID, d.deployment_config.ID,
		d.application.PreRunCommand, "pre run")
}

func (s *DeployService) PostRunCommands(d DeployerConfig) error {
	return s.runCommands(d.application.ID, d.deployment_config.ID,
		d.application.PostRunCommand, "post run")
}
