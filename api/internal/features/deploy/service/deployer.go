package service

import (
	"fmt"
	// "github.com/docker/docker/api/types/image"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type DeployerConfig struct {
	application       shared_types.Application
	deployment        *shared_types.DeploymentRequestConfig
	userID            uuid.UUID
	contextPath       string
	appStatus         shared_types.ApplicationDeploymentStatus
	deployment_config *shared_types.ApplicationDeployment
}

func (s *DeployService) runCommands(applicationID uuid.UUID, deploymentConfigID uuid.UUID,
	command string, commandType string) error {
	s.addLog(applicationID, fmt.Sprintf("Running %s commands %v", commandType, command), deploymentConfigID)

	if command == "" {
		return nil
	}

	client := ssh.NewSSH()
	output, err := client.RunCommand(command)
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

// Deployer deploys an application using the specified build pack.
//
// The build pack is determined by the Application field of the DeployerConfig
// argument. If the build pack is Dockerfile, the function calls
// handleDockerfileDeployment. If the build pack is Docker Compose, the function
// calls handleDockerComposeDeployment. If the build pack is neither of these,
// the function returns ErrInvalidBuildPack.
func (s *DeployService) Deployer(d DeployerConfig) error {
	s.addLog(d.application.ID, fmt.Sprintf(types.LogDeploymentBuildPack, d.application.BuildPack), d.deployment_config.ID)
	s.PrerunCommands(d)

	var err error
	switch d.application.BuildPack {
	case shared_types.DockerFile:
		err = s.handleDockerfileDeployment(d)
	case shared_types.DockerCompose:
		err = s.handleDockerComposeDeployment(d)
	default:
		return types.ErrInvalidBuildPack
	}

	if err != nil {
		return err
	}

	return s.PostRunCommands(d)
}

// handleDockerfileDeployment processes Dockerfile-based deployments
func (s *DeployService) handleDockerfileDeployment(d DeployerConfig) error {
	s.addLog(d.application.ID, types.LogUsingDockerfileStrategy, d.deployment_config.ID)
	s.addLog(d.application.ID, fmt.Sprintf(types.LogBuildContextPath, d.contextPath), d.deployment_config.ID)
	if err := s.buildAndRunDockerImage(d); err != nil {
		return err
	}

	return nil
}

// handleDockerComposeDeployment processes Docker Compose-based deployments
func (s *DeployService) handleDockerComposeDeployment(d DeployerConfig) error {
	s.addLog(d.application.ID, types.LogUsingDockerComposeStrategy, d.deployment_config.ID)
	s.addLog(d.application.ID, types.LogDockerComposeNotImplemented, d.deployment_config.ID)
	return types.ErrDockerComposeNotImplemented
}

// buildAndRunDockerImage handles the Docker image building and running process
func (s *DeployService) buildAndRunDockerImage(d DeployerConfig) error {
	_, err := s.buildImageFromDockerfile(d)
	if err != nil {
		s.addLog(d.application.ID, fmt.Sprintf(types.LogFailedToBuildDockerImage, err.Error()), d.deployment_config.ID)
		return fmt.Errorf("%w: %v", types.ErrBuildDockerImage, err)
	}

	s.logger.Log(logger.Info, types.LogDockerImageBuiltSuccessfully, d.application.Name)
	s.addLog(d.application.ID, types.LogDockerImageBuiltSuccessfully, d.deployment_config.ID)
	containerID, err := s.AtomicUpdateContainer(d)
	if err != nil {
		s.addLog(d.application.ID, fmt.Sprintf(types.LogFailedToRunDockerImage, err.Error()), d.deployment_config.ID)
		return fmt.Errorf("%w: %v", types.ErrRunDockerImage, err)
	}

	s.addLog(d.application.ID, fmt.Sprintf(types.LogContainerRunning, containerID), d.deployment_config.ID)
	s.addLog(d.application.ID, fmt.Sprintf(types.LogApplicationExposed, d.application.Port), d.deployment_config.ID)

	return nil
}
