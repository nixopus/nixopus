package dashboard

import (
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (p *OrgPoller) getDeployments() {
	if p.deployService == nil || p.organizationID == "" {
		p.log.Log(logger.Error, "Deploy service or organization ID not set", "")
		p.broadcastError("Deploy service or organization ID not configured", GetDeployments)
		return
	}

	var deployments []shared_types.ApplicationDeployment
	var err error

	if p.serverID != "" {
		deployments, err = p.deployService.GetLatestDeploymentsByServer(p.organizationID, p.serverID, 5)
	} else {
		deployments, err = p.deployService.GetLatestDeployments(p.organizationID, 5)
	}
	if err != nil {
		p.log.Log(logger.Error, "Failed to get deployments", err.Error())
		p.broadcastError(err.Error(), GetDeployments)
		return
	}

	p.broadcast(string(GetDeployments), deployments)
}
