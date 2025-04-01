package dashboard

import (
	"sort"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (m *DashboardMonitor) GetContainers() {
	containers, err := m.dockerService.ListAllContainers()
	if err != nil {
		m.log.Log(logger.Error, "Failed to get containers", err.Error())
		m.BroadcastError(err.Error(), GetContainers)
		return
	}

	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Created > containers[j].Created
	})

	if len(containers) > 4 {
		containers = containers[:4]
	}

	m.Broadcast(string(GetContainers), containers)
}
