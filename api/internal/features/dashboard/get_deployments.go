package dashboard

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (m *DashboardMonitor) GetDeployments() {
	if m.deployService == nil || m.organizationID == "" {
		m.log.Log(logger.Error, "Deploy service or organization ID not set", "")
		m.BroadcastError("Deploy service or organization ID not configured", GetDeployments)
		return
	}

	deployments, err := m.deployService.GetLatestDeployments(m.organizationID, 5)
	if err != nil {
		m.log.Log(logger.Error, "Failed to get deployments", err.Error())
		m.BroadcastError(err.Error(), GetDeployments)
		return
	}

	m.Broadcast(string(GetDeployments), deployments)
}
