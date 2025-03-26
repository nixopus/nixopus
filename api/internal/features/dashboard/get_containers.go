package dashboard

import (
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (m *DashboardMonitor) GetContainers() {
	containers, err := m.dockerService.ListAllContainers()
	if err != nil {
		m.log.Log(logger.Error, "Failed to get containers", err.Error())
		m.BroadcastError(err.Error(), GetContainers)
		return
	}
	m.Broadcast(string(GetContainers), containers)
}
