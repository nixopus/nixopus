package dashboard

import (
	"sort"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (p *OrgPoller) getContainers() {
	containers, err := p.dockerService.ListAllContainers()
	if err != nil {
		p.log.Log(logger.Error, "Failed to get containers", err.Error())
		p.broadcastError(err.Error(), GetContainers)
		return
	}

	sort.Slice(containers, func(i, j int) bool {
		return containers[i].Created > containers[j].Created
	})

	if len(containers) > 4 {
		containers = containers[:4]
	}

	p.broadcast(string(GetContainers), containers)
}
