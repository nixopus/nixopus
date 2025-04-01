package service

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
)

func (s *DeployService) handleDockerComposeDeployment(d DeployerConfig) error {
	s.addLog(d.application.ID, types.LogUsingDockerComposeStrategy, d.deployment_config.ID)

	composeContextPath := d.contextPath
	if d.application.BasePath != "" && d.application.BasePath != "/" {
		composeContextPath = filepath.Join(d.contextPath, d.application.BasePath)
	}

	composeFilePath := "docker-compose.yml"
	if d.application.DockerfilePath != "" {
		composeFilePath = d.application.DockerfilePath
		if strings.HasPrefix(composeFilePath, "/") {
			composeFilePath = composeFilePath[1:]
		}
	}
	absComposePath := filepath.Join(composeContextPath, composeFilePath)

	s.addLog(d.application.ID, fmt.Sprintf(types.LogBuildContextPath, composeContextPath), d.deployment_config.ID)
	s.addLog(d.application.ID, fmt.Sprintf("Using docker-compose file: %s", absComposePath), d.deployment_config.ID)

	envVars := make(map[string]string)
	for k, v := range GetMapFromString(d.application.EnvironmentVariables) {
		envVars[k] = v
	}
	for k, v := range GetMapFromString(d.application.BuildVariables) {
		envVars[k] = v
	}
	s.addLog(d.application.ID, "Building Docker Compose services...", d.deployment_config.ID)
	if err := s.dockerRepo.ComposeBuild(absComposePath, envVars); err != nil {
		s.addLog(d.application.ID, fmt.Sprintf("Error building Docker Compose services: %v", err), d.deployment_config.ID)
		return fmt.Errorf("%w: %v", types.ErrDockerComposeCommandFailed, err)
	}

	s.addLog(d.application.ID, "Starting Docker Compose services...", d.deployment_config.ID)
	if err := s.dockerRepo.ComposeUp(absComposePath, envVars); err != nil {
		s.addLog(d.application.ID, fmt.Sprintf("Error starting Docker Compose services: %v", err), d.deployment_config.ID)
		return fmt.Errorf("%w: %v", types.ErrDockerComposeCommandFailed, err)
	}

	s.addLog(d.application.ID, "Docker Compose deployment completed successfully", d.deployment_config.ID)
	return nil
}
