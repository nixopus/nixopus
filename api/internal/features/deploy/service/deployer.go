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

func (s *DeployService) PrerunCommands(d DeployerConfig) error {
	s.addLog(d.application.ID, fmt.Sprintf("Running pre run commands %v", d.application.PreRunCommand), d.deployment_config.ID)

	client := ssh.NewSSH()
	output, err := client.RunCommand(d.deployment_config.Application.PreRunCommand)
	if err != nil {
		s.addLog(d.application.ID, fmt.Sprintf("Error while running pre run command %v", output), d.deployment_config.ID)
		return err
	}
	if output != "" {
		s.addLog(d.application.ID, fmt.Sprintf("Pre run command Resulted in %v", output), d.deployment_config.ID)
	}
	return nil
}

func (s *DeployService) PostRunCommands(d DeployerConfig) error {
	s.addLog(d.application.ID, fmt.Sprintf("Running post run commands %v", d.application.PostRunCommand), d.deployment_config.ID)
	client := ssh.NewSSH()
	output, err := client.RunCommand(d.deployment_config.Application.PreRunCommand)
	if err != nil {
		s.addLog(d.application.ID, fmt.Sprintf("Error while running post run command %v", output), d.deployment_config.ID)
		return err
	}
	if output != "" {
		s.addLog(d.application.ID, fmt.Sprintf("Post run command Resulted in %v", output), d.deployment_config.ID)
	}
	return nil
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
