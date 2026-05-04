package service

import (
	"errors"
	"testing"

	"github.com/docker/docker/api/types/container"
	apinetwork "github.com/docker/docker/api/types/network"
	containertypes "github.com/nixopus/nixopus/api/internal/features/container/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/stretchr/testify/require"
)

func inspectForList(cid string) container.InspectResponse {
	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:      cid,
			Created: "2024-06-01T00:00:00Z",
			HostConfig: &container.HostConfig{
				Resources: container.Resources{
					Memory:     100,
					MemorySwap: 200,
					CPUShares:  2,
				},
			},
		},
		Config: &container.Config{Cmd: []string{"/entry"}},
		NetworkSettings: &container.NetworkSettings{
			DefaultNetworkSettings: container.DefaultNetworkSettings{IPAddress: "10.0.0.1"},
			Networks:               map[string]*apinetwork.EndpointSettings{"bridge": {}},
		},
	}
}

func TestListContainers_listDockerError(t *testing.T) {
	t.Parallel()
	stub := &stubDockerRepository{}
	stub.listContainers = func(container.ListOptions) ([]container.Summary, error) {
		return nil, errors.New("list failed")
	}
	_, err := ListContainers(stub, logger.NewLogger(), containertypes.ContainerListParams{Page: 1, PageSize: 10})
	require.Error(t, err)
}

func TestListContainers_buildDockerFilters_captured(t *testing.T) {
	t.Parallel()
	var got container.ListOptions
	stub := &stubDockerRepository{}
	stub.listContainers = func(opts container.ListOptions) ([]container.Summary, error) {
		got = opts
		return nil, errors.New("stop")
	}
	_, _ = ListContainers(stub, logger.NewLogger(), containertypes.ContainerListParams{
		Status: "running",
		Name:   "web",
		Image:  "nginx",
	})

	require.True(t, got.All)
	require.Equal(t, []string{"running"}, got.Filters.Get("status"))
	require.Equal(t, []string{"web"}, got.Filters.Get("name"))
	require.Equal(t, []string{"nginx"}, got.Filters.Get("ancestor"))
}

func TestListContainers_sortGroupingPaginationInspectSkip(t *testing.T) {
	t.Parallel()
	stub := &stubDockerRepository{}
	stub.listContainers = func(container.ListOptions) ([]container.Summary, error) {
		return []container.Summary{
			{
				ID:    "gid-b",
				Names: []string{"/bbb"},
				Image: "i2",
				Labels: map[string]string{
					"com.application.id":   "b1",
					"com.application.name": "Beta App",
				},
				Status:  "Up",
				State:   "running",
				Created: 200,
			},
			{
				ID:    "gid-a",
				Names: []string{"/aaa"},
				Image: "i1",
				Labels: map[string]string{
					"com.application.id":   "a1",
					"com.application.name": "Alpha App",
				},
				Status:  "Up",
				State:   "running",
				Created: 100,
			},
			{
				ID:      "solo",
				Names:   []string{"/x"},
				Image:   "soloimg",
				Labels:  nil,
				Status:  "Exited",
				State:   "exited",
				Created: 50,
				Ports:   []container.Port{{PrivatePort: 443, Type: "tcp"}},
			},
			{
				ID:      "badinspect",
				Names:   []string{"/badinspect"},
				Image:   "z",
				Status:  "Up",
				Created: 300,
			},
		}, nil
	}
	stub.getContainerByID = func(id string) (container.InspectResponse, error) {
		if id == "badinspect" {
			return container.InspectResponse{}, errors.New("inspect fail")
		}
		return inspectForList(id), nil
	}

	resp, err := ListContainers(stub, logger.NewLogger(), containertypes.ContainerListParams{
		Page: 1, PageSize: 1, SortBy: "name", SortOrder: "asc",
	})
	require.NoError(t, err)
	require.Equal(t, 3, resp.Data.TotalCount)
	require.Len(t, resp.Data.Ungrouped, 1)
	require.Equal(t, "solo", resp.Data.Ungrouped[0].ID)
	require.Equal(t, 2, resp.Data.GroupCount)
	require.Len(t, resp.Data.Groups, 1)

	resp2, err := ListContainers(stub, logger.NewLogger(), containertypes.ContainerListParams{
		Page: 2, PageSize: 1, SortBy: "name", SortOrder: "asc",
	})
	require.NoError(t, err)
	require.Len(t, resp2.Data.Groups, 1)

	_, err = ListContainers(stub, logger.NewLogger(), containertypes.ContainerListParams{
		SortOrder: "desc", Page: 1, PageSize: 99,
	})
	require.NoError(t, err)

	_, err = ListContainers(stub, logger.NewLogger(), containertypes.ContainerListParams{
		SortBy: "status", SortOrder: "desc",
	})
	require.NoError(t, err)

	_, err = ListContainers(stub, logger.NewLogger(), containertypes.ContainerListParams{
		SortBy: "created", SortOrder: "desc",
	})
	require.NoError(t, err)

	respFar, err := ListContainers(stub, logger.NewLogger(), containertypes.ContainerListParams{
		SortOrder: "asc", Page: 100, PageSize: 10,
	})
	require.NoError(t, err)
	require.Empty(t, respFar.Data.Groups)
}

func TestSummarizeContainers_edgeNames(t *testing.T) {
	t.Parallel()
	row := summarizeContainers([]container.Summary{
		{ID: "e1", Names: []string{}, Image: "i", Created: 1},
		{ID: "e2", Names: []string{"/"}, Image: "i", Created: 2},
		{ID: "e3", Names: []string{"/svc"}, Image: "i", Created: 3},
	})
	require.Equal(t, "", row[0].Name)
	require.Equal(t, "/", row[1].Name)
	require.Equal(t, "svc", row[2].Name)
}

func TestApplySearchFilter(t *testing.T) {
	t.Parallel()
	rows := []containertypes.ContainerListRow{{Name: "foo", Image: "x", Status: "Up"}}
	out := applySearchFilter(rows, containertypes.ContainerListParams{})
	require.Len(t, out, 1)
	out = applySearchFilter(rows, containertypes.ContainerListParams{Search: "FOO"})
	require.Len(t, out, 1)
	require.Empty(t, applySearchFilter(rows, containertypes.ContainerListParams{Search: "nomatch"}))
}
