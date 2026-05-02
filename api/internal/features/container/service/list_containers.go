package service

import (
	"sort"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	container_types "github.com/nixopus/nixopus/api/internal/features/container/types"
	"github.com/nixopus/nixopus/api/internal/features/deploy/docker"
	"github.com/nixopus/nixopus/api/internal/features/logger"
)

// ListContainers retrieves containers grouped by application, with filtering, sorting, and pagination of groups.
func ListContainers(
	dockerService docker.DockerRepository,
	l logger.Logger,
	params container_types.ContainerListParams,
) (container_types.ListContainersResponse, error) {
	containers, err := dockerService.ListContainers(container.ListOptions{
		All:     true,
		Filters: buildDockerFilters(params),
	})
	if err != nil {
		l.Log(logger.Error, err.Error(), "")
		return container_types.ListContainersResponse{}, err
	}

	rows := summarizeContainers(containers)
	filteredRows := applySearchFilter(rows, params)
	sortedRows := applySort(filteredRows, params)

	groups, ungrouped := groupContainersByApplication(sortedRows, containers, dockerService)

	sort.SliceStable(groups, func(i, j int) bool {
		if params.SortOrder == "desc" {
			return groups[i].ApplicationName > groups[j].ApplicationName
		}
		return groups[i].ApplicationName < groups[j].ApplicationName
	})

	totalGroupCount := len(groups)
	start := (params.Page - 1) * params.PageSize
	if start > totalGroupCount {
		start = totalGroupCount
	}
	end := start + params.PageSize
	if end > totalGroupCount {
		end = totalGroupCount
	}
	paginatedGroups := groups[start:end]

	totalContainerCount := 0
	for _, group := range groups {
		totalContainerCount += len(group.Containers)
	}
	totalContainerCount += len(ungrouped)

	paginatedUngrouped := ungrouped

	return container_types.ListContainersResponse{
		Status:  "success",
		Message: "Containers fetched successfully",
		Data: container_types.ListContainersResponseData{
			Groups:     paginatedGroups,
			Ungrouped:  paginatedUngrouped,
			TotalCount: totalContainerCount,
			GroupCount: totalGroupCount,
			Page:       params.Page,
			PageSize:   params.PageSize,
			SortBy:     params.SortBy,
			SortOrder:  params.SortOrder,
			Search:     params.Search,
			Status:     params.Status,
			Name:       params.Name,
			Image:      params.Image,
		},
	}, nil
}

func buildDockerFilters(p container_types.ContainerListParams) filters.Args {
	f := filters.NewArgs()
	if p.Status != "" {
		f.Add("status", p.Status)
	}
	if p.Name != "" {
		f.Add("name", p.Name)
	}
	if p.Image != "" {
		f.Add("ancestor", p.Image)
	}
	return f
}

func summarizeContainers(summaries []container.Summary) []container_types.ContainerListRow {
	rows := make([]container_types.ContainerListRow, 0, len(summaries))
	for _, csum := range summaries {
		name := ""
		if len(csum.Names) > 0 {
			n := csum.Names[0]
			if len(n) > 1 {
				name = n[1:]
			} else {
				name = n
			}
		}
		rows = append(rows, container_types.ContainerListRow{
			ID:      csum.ID,
			Name:    name,
			Image:   csum.Image,
			Status:  csum.Status,
			State:   csum.State,
			Created: csum.Created,
			Labels:  csum.Labels,
		})
	}
	return rows
}

func applySearchFilter(rows []container_types.ContainerListRow, p container_types.ContainerListParams) []container_types.ContainerListRow {
	if p.Search != "" {
		lower := strings.ToLower(p.Search)
		filtered := make([]container_types.ContainerListRow, 0, len(rows))
		for _, r := range rows {
			if strings.Contains(strings.ToLower(r.Name), lower) ||
				strings.Contains(strings.ToLower(r.Image), lower) ||
				strings.Contains(strings.ToLower(r.Status), lower) {
				filtered = append(filtered, r)
			}
		}
		return filtered
	}
	return rows
}

func applySort(rows []container_types.ContainerListRow, p container_types.ContainerListParams) []container_types.ContainerListRow {
	sort.SliceStable(rows, func(i, j int) bool {
		switch p.SortBy {
		case "status":
			a := strings.ToLower(rows[i].Status)
			b := strings.ToLower(rows[j].Status)
			if p.SortOrder == "desc" {
				return a > b
			}
			return a < b
		case "name":
			a := strings.ToLower(rows[i].Name)
			b := strings.ToLower(rows[j].Name)
			if p.SortOrder == "desc" {
				return a > b
			}
			return a < b
		default:
			ai := rows[i].Created
			aj := rows[j].Created
			if p.SortOrder == "desc" {
				return ai > aj
			}
			return ai < aj
		}
	})
	return rows
}

func groupContainersByApplication(
	rows []container_types.ContainerListRow,
	summaries []container.Summary,
	dockerService interface {
		GetContainerById(id string) (container.InspectResponse, error)
	},
) ([]container_types.ContainerGroup, []container_types.Container) {
	groupsMap := make(map[string]*container_types.ContainerGroup)
	ungrouped := make([]container_types.Container, 0)

	summaryMap := make(map[string]container.Summary)
	for _, s := range summaries {
		summaryMap[s.ID] = s
	}

	for _, row := range rows {
		applicationID := ""
		applicationName := "Unknown Application"
		if row.Labels != nil {
			if id, ok := row.Labels["com.application.id"]; ok {
				applicationID = id
			}
			if name, ok := row.Labels["com.application.name"]; ok {
				applicationName = name
			}
		}

		info, err := dockerService.GetContainerById(row.ID)
		if err != nil {
			continue
		}

		containerData := container_types.Container{
			ID:        row.ID,
			Name:      row.Name,
			Image:     row.Image,
			Status:    row.Status,
			State:     row.State,
			Created:   info.Created,
			Labels:    row.Labels,
			Command:   "",
			IPAddress: info.NetworkSettings.IPAddress,
			HostConfig: container_types.HostConfig{
				Memory:     info.HostConfig.Memory,
				MemorySwap: info.HostConfig.MemorySwap,
				CPUShares:  info.HostConfig.CPUShares,
			},
		}

		if info.Config != nil && info.Config.Cmd != nil && len(info.Config.Cmd) > 0 {
			containerData.Command = info.Config.Cmd[0]
		}

		if s, ok := summaryMap[row.ID]; ok {
			for _, p := range s.Ports {
				containerData.Ports = append(containerData.Ports, container_types.Port{
					PrivatePort: int(p.PrivatePort),
					PublicPort:  int(p.PublicPort),
					Type:        p.Type,
				})
			}
		}

		for _, m := range info.Mounts {
			containerData.Mounts = append(containerData.Mounts, container_types.Mount{
				Type:        string(m.Type),
				Source:      m.Source,
				Destination: m.Destination,
				Mode:        m.Mode,
			})
		}

		for name, network := range info.NetworkSettings.Networks {
			containerData.Networks = append(containerData.Networks, container_types.Network{
				Name:       name,
				IPAddress:  network.IPAddress,
				Gateway:    network.Gateway,
				MacAddress: network.MacAddress,
				Aliases:    network.Aliases,
			})
		}

		if applicationID != "" {
			if _, exists := groupsMap[applicationID]; !exists {
				groupsMap[applicationID] = &container_types.ContainerGroup{
					ApplicationID:   applicationID,
					ApplicationName: applicationName,
					Containers:      make([]container_types.Container, 0),
				}
			}
			groupsMap[applicationID].Containers = append(groupsMap[applicationID].Containers, containerData)
		} else {
			ungrouped = append(ungrouped, containerData)
		}
	}

	groups := make([]container_types.ContainerGroup, 0, len(groupsMap))
	for _, group := range groupsMap {
		groups = append(groups, *group)
	}

	return groups, ungrouped
}
