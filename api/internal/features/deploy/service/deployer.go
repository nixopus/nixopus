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

// Deployer handles deployment processes based on the specified build pack type.
//
// This method logs the start of the deployment process and executes different
// deployment strategies depending on the build pack type specified in the
// CreateDeploymentRequest. For Dockerfile build packs, it builds a Docker image
// using the provided build variables and environment variables, then runs the
// Docker image. For DockerCompose build packs, it logs the intended operation
// and returns without performing any actions.
//
// Parameters:
//
//	applicationID - the UUID of the application.
//	deployment - a pointer to the CreateDeploymentRequest containing the deployment details.
//	userID - the UUID of the user initiating the deployment.
//	contextPath - the path to the build context directory.
//
// Returns:
//
//	error - an error if the deployment process fails at any step, otherwise nil.
func (s *DeployService) Deployer(d DeployerConfig) error {
	s.logger.Log(logger.Info, "Creating deployment", d.contextPath)
	s.addLog(d.applicationID, fmt.Sprintf("Starting deployment process for build pack: %s", d.deployment.BuildPack), d.deployment_config.ID)

	switch d.deployment.BuildPack {
	case shared_types.DockerFile:
		s.addLog(d.applicationID, "Using Dockerfile build strategy", d.deployment_config.ID)
		s.logger.Log(logger.Info, "Dockerfile building", "")

		buildArgs := make(map[string]*string)
		for k, v := range d.deployment.BuildVariables {
			value := v
			buildArgs[k] = &value
		}
		s.addLog(d.applicationID, fmt.Sprintf("Using %d build arguments", len(buildArgs)), d.deployment_config.ID)

		labels := make(map[string]string)
		for k, v := range d.deployment.EnvironmentVariables {
			labels[k] = v
		}

		dockerfilePath := "Dockerfile"

		s.logger.Log(logger.Info, "Build context path", d.contextPath)
		s.addLog(d.applicationID, fmt.Sprintf("Build context path: %s", d.contextPath), d.deployment_config.ID)
		s.logger.Log(logger.Info, "Using Dockerfile", dockerfilePath)

		build_config := BuildImageFromDockerFile{
			d.applicationID,
			d.contextPath,
			dockerfilePath,
			false,
			buildArgs,
			labels,
			d.deployment.Name,
			d.statusID,
			d.deployment_config,
		}
		_, err := s.buildImageFromDockerfile(build_config)
		if err != nil {
			s.addLog(d.applicationID, fmt.Sprintf("Failed to build Docker image: %s", err.Error()), d.deployment_config.ID)
			return fmt.Errorf("failed to build Docker image: %w", err)
		}

		s.logger.Log(logger.Info, "Dockerfile built successfully", d.deployment.Name)
		s.addLog(d.applicationID, "Docker image built successfully", d.deployment_config.ID)

		run_image_config := RunImageConfig{
			d.applicationID,
			d.deployment.Name,
			d.deployment.EnvironmentVariables,
			fmt.Sprintf("%d", d.deployment.Port),
			d.statusID,
			d.deployment_config,
		}

		containerID, err := s.RunImage(run_image_config)
		if err != nil {
			s.addLog(d.applicationID, fmt.Sprintf("Failed to run Docker image: %s", err.Error()), d.deployment_config.ID)
			return fmt.Errorf("failed to run Docker image: %w", err)
		}

		s.addLog(d.applicationID, fmt.Sprintf("Container is running with ID: %s", containerID), d.deployment_config.ID)
		s.addLog(d.applicationID, fmt.Sprintf("Application exposed on port: %d", d.deployment.Port), d.deployment_config.ID)

	case shared_types.DockerCompose:
		s.logger.Log(logger.Info, "Docker compose building", "")
		s.addLog(d.applicationID, "Docker Compose deployment strategy selected", d.deployment_config.ID)
		s.addLog(d.applicationID, "Docker Compose deployment not implemented yet", d.deployment_config.ID)
		return nil
	}

	return nil
}
