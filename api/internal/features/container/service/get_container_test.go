package service

import (
	"errors"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	apinetwork "github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/stretchr/testify/require"
)

func TestGetContainer_error(t *testing.T) {
	t.Parallel()
	stub := &stubDockerRepository{}
	stub.getContainerByID = func(string) (container.InspectResponse, error) {
		return container.InspectResponse{}, errors.New("no such container")
	}
	_, err := GetContainer(stub, logger.NewLogger(), "deadbeef")
	require.Error(t, err)
}

func TestGetContainer_fullShape(t *testing.T) {
	t.Parallel()
	port, err := nat.NewPort("tcp", "8080")
	require.NoError(t, err)

	stub := &stubDockerRepository{}
	stub.getContainerByID = func(id string) (container.InspectResponse, error) {
		require.Equal(t, "abc", id)

		ns := &container.NetworkSettings{}
		ns.Ports = nat.PortMap{
			port: {{HostPort: "18080"}},
		}
		ns.Networks = map[string]*apinetwork.EndpointSettings{
			"n1": {IPAddress: "172.18.0.2", Gateway: "172.18.0.1", Aliases: nil},
		}
		ns.IPAddress = "10.11.12.13"

		return container.InspectResponse{
			ContainerJSONBase: &container.ContainerJSONBase{
				ID:      "inspect-id",
				Created: "2020-01-01",
				Name:    "/fancy",
				State: &container.State{
					Status: "running",
				},
				HostConfig: &container.HostConfig{
					Resources: container.Resources{
						Memory:     4096,
						MemorySwap: 8192,
						CPUShares:  10,
					},
				},
			},
			Config: &container.Config{
				Image:  "nginx:1",
				Labels: map[string]string{"k": "v"},
				Cmd:    []string{"/sbin/init", "--help"},
			},
			NetworkSettings: ns,
			Mounts: []container.MountPoint{
				{Type: mount.TypeBind, Source: "/a", Destination: "/b", Mode: "rw"},
			},
		}, nil
	}

	got, err := GetContainer(stub, logger.NewLogger(), "abc")
	require.NoError(t, err)

	require.Equal(t, "inspect-id", got.ID)
	require.Equal(t, "fancy", got.Name)
	require.Equal(t, "nginx:1", got.Image)
	require.Equal(t, "/sbin/init", got.Command)
	require.Equal(t, map[string]string{"k": "v"}, got.Labels)
	require.Equal(t, "2020-01-01", got.Created)
	require.Len(t, got.Ports, 1)
	require.Equal(t, 8080, got.Ports[0].PrivatePort)
	require.Equal(t, 18080, got.Ports[0].PublicPort)
	require.Equal(t, "tcp", got.Ports[0].Type)
	require.Len(t, got.Networks, 1)
	require.Equal(t, "n1", got.Networks[0].Name)
	require.Empty(t, got.Networks[0].Aliases)
	require.Equal(t, "10.11.12.13", got.IPAddress)
	require.Equal(t, int64(4096), got.HostConfig.Memory)

	require.Len(t, got.Mounts, 1)
	require.Equal(t, "bind", got.Mounts[0].Type)
}

func TestGetContainer_minimalOptionalNil(t *testing.T) {
	t.Parallel()
	stub := &stubDockerRepository{}
	stub.getContainerByID = func(string) (container.InspectResponse, error) {
		return container.InspectResponse{
			ContainerJSONBase: &container.ContainerJSONBase{
				ID:      "x",
				Created: "t",
				Name:    "no-leading-slash",
			},
			NetworkSettings: &container.NetworkSettings{},
		}, nil
	}

	got, err := GetContainer(stub, logger.NewLogger(), "x")
	require.NoError(t, err)

	require.Empty(t, got.Image)
	require.Empty(t, got.Ports)
	require.Empty(t, got.Networks)
	require.Equal(t, "x", got.ID)
	require.Equal(t, "t", got.Created)
	require.Equal(t, "o-leading-slash", got.Name)
}
