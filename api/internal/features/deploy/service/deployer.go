package service

import (
	"fmt"
	// "github.com/docker/docker/api/types/image"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type DeployerConfig struct {
	applicationID     uuid.UUID
	deployment        *types.CreateDeploymentRequest
	userID            uuid.UUID
	contextPath       string
	statusID          uuid.UUID
	deployment_config *shared_types.ApplicationDeployment
}

// Deployer starts the deployment process using the specified build pack.
//
// Parameters:
//
//	d - a DeployerConfig struct containing the application ID, deployment request,
//	    user ID, context path, status ID, and deployment configuration.
//
// Returns:
//
//	error - an error if the deployment process fails at any step, otherwise nil.
func (s *DeployService) Deployer(d DeployerConfig) error {
	s.addLog(d.applicationID, fmt.Sprintf(types.LogDeploymentBuildPack, d.deployment.BuildPack), d.deployment_config.ID)

	switch d.deployment.BuildPack {
	case shared_types.DockerFile:
		return s.handleDockerfileDeployment(d)
	case shared_types.DockerCompose:
		return s.handleDockerComposeDeployment(d)
	default:
		return types.ErrInvalidBuildPack
	}
}

// handleDockerfileDeployment processes Dockerfile-based deployments
func (s *DeployService) handleDockerfileDeployment(d DeployerConfig) error {
	s.addLog(d.applicationID, types.LogUsingDockerfileStrategy, d.deployment_config.ID)
	buildArgs := s.prepareBuildArgs(d)
	labels := s.prepareEnvironmentVariables(d)
	dockerfilePath := "Dockerfile"
	s.addLog(d.applicationID, fmt.Sprintf(types.LogBuildContextPath, d.contextPath), d.deployment_config.ID)
	s.addLog(d.applicationID, fmt.Sprintf(types.LogUsingBuildArgs, len(buildArgs)), d.deployment_config.ID)

	if err := s.buildAndRunDockerImage(d, buildArgs, labels, dockerfilePath); err != nil {
		return err
	}

	return nil
}

// handleDockerComposeDeployment processes Docker Compose-based deployments
func (s *DeployService) handleDockerComposeDeployment(d DeployerConfig) error {
	s.addLog(d.applicationID, types.LogUsingDockerComposeStrategy, d.deployment_config.ID)
	s.addLog(d.applicationID, types.LogDockerComposeNotImplemented, d.deployment_config.ID)
	return types.ErrDockerComposeNotImplemented
}

// prepareBuildArgs extracts build variables from the deployment request
func (s *DeployService) prepareBuildArgs(d DeployerConfig) map[string]*string {
	buildArgs := make(map[string]*string)
	for k, v := range d.deployment.BuildVariables {
		value := v
		buildArgs[k] = &value
	}
	return buildArgs
}

// prepareEnvironmentVariables extracts environment variables from the deployment request
func (s *DeployService) prepareEnvironmentVariables(d DeployerConfig) map[string]string {
	labels := make(map[string]string)
	for k, v := range d.deployment.EnvironmentVariables {
		labels[k] = v
	}
	return labels
}

// buildAndRunDockerImage handles the Docker image building and running process
func (s *DeployService) buildAndRunDockerImage(d DeployerConfig, buildArgs map[string]*string, labels map[string]string, dockerfilePath string) error {
	buildConfig := BuildImageFromDockerFile{
		applicationID:     d.applicationID,
		contextPath:       d.contextPath,
		dockerfile:        dockerfilePath,
		force:             false,
		buildArgs:         buildArgs,
		labels:            labels,
		image_name:        d.deployment.Name,
		statusID:          d.statusID,
		deployment_config: d.deployment_config,
	}

	_, err := s.buildImageFromDockerfile(buildConfig)
	if err != nil {
		s.addLog(d.applicationID, fmt.Sprintf(types.LogFailedToBuildDockerImage, err.Error()), d.deployment_config.ID)
		return fmt.Errorf("%w: %v", types.ErrBuildDockerImage, err)
	}

	s.logger.Log(logger.Info, types.LogDockerImageBuiltSuccessfully, d.deployment.Name)
	s.addLog(d.applicationID, types.LogDockerImageBuiltSuccessfully, d.deployment_config.ID)

	runImageConfig := RunImageConfig{
		applicationID:         d.applicationID,
		imageName:             d.deployment.Name,
		environment_variables: d.deployment.EnvironmentVariables,
		port_str:              fmt.Sprintf("%d", d.deployment.Port),
		statusID:              d.statusID,
		deployment_config:     d.deployment_config,
	}

	containerID, err := s.RunImage(runImageConfig)
	if err != nil {
		s.addLog(d.applicationID, fmt.Sprintf(types.LogFailedToRunDockerImage, err.Error()), d.deployment_config.ID)
		return fmt.Errorf("%w: %v", types.ErrRunDockerImage, err)
	}

	s.addLog(d.applicationID, fmt.Sprintf(types.LogContainerRunning, containerID), d.deployment_config.ID)
	s.addLog(d.applicationID, fmt.Sprintf(types.LogApplicationExposed, d.deployment.Port), d.deployment_config.ID)

	return nil
}
