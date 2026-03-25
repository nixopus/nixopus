package dashboard

import (
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func (p *OrgPoller) getDeployments() {
	if p.deployService == nil || p.organizationID == "" {
		p.log.Log(logger.Error, "Deploy service or organization ID not set", "")
		p.broadcastError("Deploy service or organization ID not configured", GetDeployments)
		return
	}

	deployments, err := p.deployService.GetLatestDeployments(p.organizationID, 5)
	if err != nil {
		p.log.Log(logger.Error, "Failed to get deployments", err.Error())
		p.broadcastError(err.Error(), GetDeployments)
		return
	}

	p.broadcast(string(GetDeployments), deployments)
}
