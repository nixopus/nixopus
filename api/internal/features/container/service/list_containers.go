package service

import (
	"sort"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// ListContainers retrieves a paginated, filtered, and sorted list of containers
func ListContainers(
	dockerService *docker.DockerService,
	l logger.Logger,
	params container_types.ContainerListParams,
) (container_types.ListContainersResponse, error) {
	// Get pre-filtered summaries from Docker
	containers, err := dockerService.ListContainers(container.ListOptions{
		All:     true,
		Filters: buildDockerFilters(params),
	})
	if err != nil {
		l.Log(logger.Error, err.Error(), "")
		return container_types.ListContainersResponse{}, err
	}

	// Build summaries, then search/sort/paginate
	rows := summarizeContainers(containers)
	pageRows, totalCount := applySearchSortPaginate(rows, params)

	// Build detailed container info for paginated results
	result := appendContainerInfo(dockerService, l, pageRows, containers)

	return container_types.ListContainersResponse{
		Status:  "success",
		Message: "Containers fetched successfully",
		Data: container_types.ListContainersResponseData{
			Containers: result,
			TotalCount: totalCount,
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

func applySearchSortPaginate(rows []container_types.ContainerListRow, p container_types.ContainerListParams) ([]container_types.ContainerListRow, int) {
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
		rows = filtered
	}

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

	totalCount := len(rows)
	start := (p.Page - 1) * p.PageSize
	if start > totalCount {
		start = totalCount
	}
	end := start + p.PageSize
	if end > totalCount {
		end = totalCount
	}
	return rows[start:end], totalCount
}

func appendContainerInfo(dockerService *docker.DockerService, l logger.Logger, pageRows []container_types.ContainerListRow, summaries []container.Summary) []container_types.Container {
	result := make([]container_types.Container, 0, len(pageRows))
	for _, r := range pageRows {
		info, err := dockerService.GetContainerById(r.ID)
		if err != nil {
			l.Log(logger.Error, "Error inspecting container", r.ID)
			continue
		}
		cd := container_types.Container{
			ID:        r.ID,
			Name:      r.Name,
			Image:     r.Image,
			Status:    r.Status,
			State:     r.State,
			Created:   info.Created,
			Labels:    r.Labels,
			Ports:     []container_types.Port{},
			Mounts:    []container_types.Mount{},
			Networks:  []container_types.Network{},
			Command:   "",
			IPAddress: info.NetworkSettings.IPAddress,
			HostConfig: container_types.HostConfig{
				Memory:     info.HostConfig.Memory,
				MemorySwap: info.HostConfig.MemorySwap,
				CPUShares:  info.HostConfig.CPUShares,
			},
		}
		if info.Config != nil && info.Config.Cmd != nil && len(info.Config.Cmd) > 0 {
			cd.Command = info.Config.Cmd[0]
		}
		for _, s := range summaries {
			if s.ID == r.ID {
				for _, p := range s.Ports {
					cd.Ports = append(cd.Ports, container_types.Port{
						PrivatePort: int(p.PrivatePort),
						PublicPort:  int(p.PublicPort),
						Type:        p.Type,
					})
				}
				break
			}
		}
		for _, m := range info.Mounts {
			cd.Mounts = append(cd.Mounts, container_types.Mount{
				Type:        string(m.Type),
				Source:      m.Source,
				Destination: m.Destination,
				Mode:        m.Mode,
			})
		}
		for name, network := range info.NetworkSettings.Networks {
			cd.Networks = append(cd.Networks, container_types.Network{
				Name:       name,
				IPAddress:  network.IPAddress,
				Gateway:    network.Gateway,
				MacAddress: network.MacAddress,
				Aliases:    network.Aliases,
			})
		}
		result = append(result, cd)
	}
	return result
}
