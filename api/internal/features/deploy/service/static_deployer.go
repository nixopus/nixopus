package service

import (
	"fmt"
	"strconv"

	"github.com/raghavyuva/nixopus-api/internal/features/deploy/proxy"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *DeployService) handleStaticDeployment(d DeployerConfig) error {
	s.addLog(d.application.ID, "Using static file deployment strategy", d.deployment_config.ID)

	caddyProxy := proxy.NewCaddy(&s.logger, d.contextPath, d.application.Domain, strconv.Itoa(443), proxy.FileServer)
	if err := caddyProxy.Serve(); err != nil {
		s.addLog(d.application.ID, fmt.Sprintf("Failed to start Caddy proxy: %v", err), d.deployment_config.ID)
		return err
	}
	s.addLog(d.application.ID, "Caddy proxy started successfully", d.deployment_config.ID)

	s.updateStatus(d.deployment_config.ID, shared_types.Deployed, d.appStatus.ID)
	s.addLog(d.application.ID, types.LogDeploymentCompletedSuccessfully, d.deployment_config.ID)
	s.addLog(d.application.ID, fmt.Sprintf("Application %s is available at %s", d.application.Name, d.application.Domain), d.deployment_config.ID)
	return nil
}
