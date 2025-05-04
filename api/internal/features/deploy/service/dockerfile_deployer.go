package service

import (
	"fmt"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/proxy"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (s *DeployService) handleDockerfileDeployment(d DeployerConfig) error {
	s.addLog(d.application.ID, types.LogUsingDockerfileStrategy, d.deployment_config.ID)
	s.addLog(d.application.ID, fmt.Sprintf(types.LogBuildContextPath, d.contextPath), d.deployment_config.ID)
	availablePort, err := s.buildAndRunDockerImage(d)
	if err != nil {
		return err
	}

	caddyProxy := proxy.NewCaddy(&s.logger, d.contextPath, d.application.Domain, availablePort, proxy.ReverseProxy)
	if err := caddyProxy.Serve(); err != nil {
		s.addLog(d.application.ID, fmt.Sprintf("Failed to start Caddy proxy: %v", err), d.deployment_config.ID)
		return err
	}
	s.addLog(d.application.ID, "Caddy proxy started successfully", d.deployment_config.ID)

	return nil
}

func (s *DeployService) buildAndRunDockerImage(d DeployerConfig) (string, error) {
	_, err := s.buildImageFromDockerfile(d)
	if err != nil {
		s.addLog(d.application.ID, fmt.Sprintf(types.LogFailedToBuildDockerImage, err.Error()), d.deployment_config.ID)
		return "", fmt.Errorf("%w: %v", types.ErrBuildDockerImage, err)
	}

	s.logger.Log(logger.Info, types.LogDockerImageBuiltSuccessfully, d.application.Name)
	s.addLog(d.application.ID, types.LogDockerImageBuiltSuccessfully, d.deployment_config.ID)
	containerID, availablePort, err := s.AtomicUpdateContainer(d)
	if err != nil {
		s.addLog(d.application.ID, fmt.Sprintf(types.LogFailedToRunDockerImage, err.Error()), d.deployment_config.ID)
		return "", fmt.Errorf("%w: %v", types.ErrRunDockerImage, err)
	}

	s.addLog(d.application.ID, fmt.Sprintf(types.LogContainerRunning, containerID), d.deployment_config.ID)
	s.addLog(d.application.ID, fmt.Sprintf(types.LogApplicationExposed, d.application.Port), d.deployment_config.ID)

	return availablePort, nil
}
