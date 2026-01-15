package controller

import (
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/go-fuego/fuego"
	containertypes "github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (c *ContainerController) ListContainers(fuegoCtx fuego.ContextNoBody) (*containertypes.ListContainersResponse, error) {
	// normalize query params
	params := parseContainerListParams(fuegoCtx.Request())

	// Get pre-filtered summaries from Docker
	containers, err := c.dockerService.ListContainers(container.ListOptions{
		All:     true,
		Filters: buildDockerFilters(params),
	})
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}
	// Build summaries, then search/sort
	rows := summarizeContainers(containers)
	filteredRows := applySearchFilter(rows, params)
	sortedRows := applySort(filteredRows, params)

	// Group containers by application ID
	groups, ungrouped := groupContainersByApplication(sortedRows, containers, c.dockerService)

	// Sort groups by application name
	sort.SliceStable(groups, func(i, j int) bool {
		if params.SortOrder == "desc" {
			return groups[i].ApplicationName > groups[j].ApplicationName
		}
		return groups[i].ApplicationName < groups[j].ApplicationName
	})

	// Paginate groups
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

	// Calculate total container count
	totalContainerCount := 0
	for _, group := range groups {
		totalContainerCount += len(group.Containers)
	}
	totalContainerCount += len(ungrouped)

	// Include ungrouped containers on every page
	// This ensures they're always visible regardless of pagination
	paginatedUngrouped := ungrouped

	return &containertypes.ListContainersResponse{
		Status:  "success",
		Message: "Containers fetched successfully",
		Data: containertypes.ListContainersResponseData{
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

func parseContainerListParams(r *http.Request) containertypes.ContainerListParams {
	q := r.URL.Query()
	pageStr := q.Get("page")
	pageSizeStr := q.Get("page_size")
	sortBy := strings.ToLower(strings.TrimSpace(q.Get("sort_by")))
	sortOrder := strings.ToLower(strings.TrimSpace(q.Get("sort_order")))

	if pageStr == "" {
		pageStr = "1"
	}
	if pageSizeStr == "" {
		pageSizeStr = "10"
	}
	if sortBy == "" {
		sortBy = "name"
	}
	if sortOrder == "" {
		sortOrder = "asc"
	}

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 {
		pageSize = 10
	}

	return containertypes.ContainerListParams{
		Page:      page,
		PageSize:  pageSize,
		Search:    strings.TrimSpace(q.Get("search")),
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Status:    strings.TrimSpace(q.Get("status")),
		Name:      strings.TrimSpace(q.Get("name")),
		Image:     strings.TrimSpace(q.Get("image")),
	}
}

func buildDockerFilters(p containertypes.ContainerListParams) filters.Args {
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

func summarizeContainers(summaries []container.Summary) []containertypes.ContainerListRow {
	rows := make([]containertypes.ContainerListRow, 0, len(summaries))
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
		rows = append(rows, containertypes.ContainerListRow{
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

func applySearchFilter(rows []containertypes.ContainerListRow, p containertypes.ContainerListParams) []containertypes.ContainerListRow {
	if p.Search != "" {
		lower := strings.ToLower(p.Search)
		filtered := make([]containertypes.ContainerListRow, 0, len(rows))
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

func applySort(rows []containertypes.ContainerListRow, p containertypes.ContainerListParams) []containertypes.ContainerListRow {
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
	rows []containertypes.ContainerListRow,
	summaries []container.Summary,
	dockerService interface {
		GetContainerById(id string) (container.InspectResponse, error)
	},
) ([]containertypes.ContainerGroup, []containertypes.Container) {
	groupsMap := make(map[string]*containertypes.ContainerGroup)
	ungrouped := make([]containertypes.Container, 0)

	// Create a map of summaries by ID for quick lookup
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

		// Get full container info
		info, err := dockerService.GetContainerById(row.ID)
		if err != nil {
			continue
		}

		containerData := containertypes.Container{
			ID:        row.ID,
			Name:      row.Name,
			Image:     row.Image,
			Status:    row.Status,
			State:     row.State,
			Created:   info.Created,
			Labels:    row.Labels,
			Command:   "",
			IPAddress: info.NetworkSettings.IPAddress,
			HostConfig: containertypes.HostConfig{
				Memory:     info.HostConfig.Memory,
				MemorySwap: info.HostConfig.MemorySwap,
				CPUShares:  info.HostConfig.CPUShares,
			},
		}

		if info.Config != nil && info.Config.Cmd != nil && len(info.Config.Cmd) > 0 {
			containerData.Command = info.Config.Cmd[0]
		}

		// Add ports from summary
		if s, ok := summaryMap[row.ID]; ok {
			for _, p := range s.Ports {
				containerData.Ports = append(containerData.Ports, containertypes.Port{
					PrivatePort: int(p.PrivatePort),
					PublicPort:  int(p.PublicPort),
					Type:        p.Type,
				})
			}
		}

		// Add mounts
		for _, m := range info.Mounts {
			containerData.Mounts = append(containerData.Mounts, containertypes.Mount{
				Type:        string(m.Type),
				Source:      m.Source,
				Destination: m.Destination,
				Mode:        m.Mode,
			})
		}

		// Add networks
		for name, network := range info.NetworkSettings.Networks {
			containerData.Networks = append(containerData.Networks, containertypes.Network{
				Name:       name,
				IPAddress:  network.IPAddress,
				Gateway:    network.Gateway,
				MacAddress: network.MacAddress,
				Aliases:    network.Aliases,
			})
		}

		if applicationID != "" {
			if _, exists := groupsMap[applicationID]; !exists {
				groupsMap[applicationID] = &containertypes.ContainerGroup{
					ApplicationID:   applicationID,
					ApplicationName: applicationName,
					Containers:      make([]containertypes.Container, 0),
				}
			}
			groupsMap[applicationID].Containers = append(groupsMap[applicationID].Containers, containerData)
		} else {
			ungrouped = append(ungrouped, containerData)
		}
	}

	// Convert map to slice
	groups := make([]containertypes.ContainerGroup, 0, len(groupsMap))
	for _, group := range groupsMap {
		groups = append(groups, *group)
	}

	return groups, ungrouped
}

func (c *ContainerController) appendContainerInfo(pageRows []containertypes.ContainerListRow, summaries []container.Summary) []containertypes.Container {
	result := make([]containertypes.Container, 0, len(pageRows))
	for _, r := range pageRows {
		info, err := c.dockerService.GetContainerById(r.ID)
		if err != nil {
			c.logger.Log(logger.Error, "Error inspecting container", r.ID)
			continue
		}
		cd := containertypes.Container{
			ID:        r.ID,
			Name:      r.Name,
			Image:     r.Image,
			Status:    r.Status,
			State:     r.State,
			Created:   info.Created,
			Labels:    r.Labels,
			Command:   "",
			IPAddress: info.NetworkSettings.IPAddress,
			HostConfig: containertypes.HostConfig{
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
					cd.Ports = append(cd.Ports, containertypes.Port{
						PrivatePort: int(p.PrivatePort),
						PublicPort:  int(p.PublicPort),
						Type:        p.Type,
					})
				}
				break
			}
		}
		for _, m := range info.Mounts {
			cd.Mounts = append(cd.Mounts, containertypes.Mount{
				Type:        string(m.Type),
				Source:      m.Source,
				Destination: m.Destination,
				Mode:        m.Mode,
			})
		}
		for name, network := range info.NetworkSettings.Networks {
			cd.Networks = append(cd.Networks, containertypes.Network{
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
