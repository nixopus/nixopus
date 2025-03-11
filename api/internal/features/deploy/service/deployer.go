package service

import (
	"fmt"
	// "github.com/docker/docker/api/types/image"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

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
func (s *DeployService) Deployer(applicationID uuid.UUID, deployment *types.CreateDeploymentRequest, userID uuid.UUID, contextPath string, statusID uuid.UUID) error {
	s.logger.Log(logger.Info, "Creating deployment", contextPath)
	s.addLog(applicationID, fmt.Sprintf("Starting deployment process for build pack: %s", deployment.BuildPack))

	switch deployment.BuildPack {
	case shared_types.DockerFile:
		s.addLog(applicationID, "Using Dockerfile build strategy")
		s.logger.Log(logger.Info, "Dockerfile building", "")

		buildArgs := make(map[string]*string)
		for k, v := range deployment.BuildVariables {
			value := v
			buildArgs[k] = &value
		}
		s.addLog(applicationID, fmt.Sprintf("Using %d build arguments", len(buildArgs)))

		labels := make(map[string]string)
		for k, v := range deployment.EnvironmentVariables {
			labels[k] = v
		}

		dockerfilePath := "Dockerfile"

		s.logger.Log(logger.Info, "Build context path", contextPath)
		s.addLog(applicationID, fmt.Sprintf("Build context path: %s", contextPath))
		s.logger.Log(logger.Info, "Using Dockerfile", dockerfilePath)

		_, err := s.buildImageFromDockerfile(
			applicationID,
			contextPath,
			dockerfilePath,
			false,
			buildArgs,
			labels,
			deployment.Name,
			statusID,
		)
		if err != nil {
			s.addLog(applicationID, fmt.Sprintf("Failed to build Docker image: %s", err.Error()))
			return fmt.Errorf("failed to build Docker image: %w", err)
		}

		s.logger.Log(logger.Info, "Dockerfile built successfully", deployment.Name)
		s.addLog(applicationID, "Docker image built successfully")

		containerID, err := s.RunImage(applicationID, deployment.Name, deployment.EnvironmentVariables, fmt.Sprintf("%d", deployment.Port), statusID)
		if err != nil {
			s.addLog(applicationID, fmt.Sprintf("Failed to run Docker image: %s", err.Error()))
			return fmt.Errorf("failed to run Docker image: %w", err)
		}

		s.addLog(applicationID, fmt.Sprintf("Container is running with ID: %s", containerID))
		s.addLog(applicationID, fmt.Sprintf("Application exposed on port: %d", deployment.Port))

	case shared_types.DockerCompose:
		s.logger.Log(logger.Info, "Docker compose building", "")
		s.addLog(applicationID, "Docker Compose deployment strategy selected")
		s.addLog(applicationID, "Docker Compose deployment not implemented yet")
		return nil
	}

	return nil
}
