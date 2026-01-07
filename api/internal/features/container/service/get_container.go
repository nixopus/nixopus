package service

import (
	"strconv"

	container_types "github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// GetContainer retrieves detailed container information and transforms it to Container type.
func GetContainer(
	dockerService *docker.DockerService,
	l logger.Logger,
	containerID string,
) (container_types.Container, error) {
	containerInfo, err := dockerService.GetContainerById(containerID)
	if err != nil {
		l.Log(logger.Error, err.Error(), "")
		return container_types.Container{}, err
	}

	containerData := container_types.Container{
		ID:      containerInfo.ID,
		Name:    containerInfo.Name,
		Image:   "",
		Status:  "",
		State:   "",
		Created: containerInfo.Created,
		Labels:  make(map[string]string),
		Command: "",
	}

	if len(containerInfo.Name) > 0 {
		containerData.Name = containerInfo.Name[1:]
	}

	if containerInfo.Config != nil {
		containerData.Image = containerInfo.Config.Image
		if containerInfo.Config.Labels != nil {
			containerData.Labels = containerInfo.Config.Labels
		}
		if len(containerInfo.Config.Cmd) > 0 {
			containerData.Command = containerInfo.Config.Cmd[0]
		}
	}

	if containerInfo.State != nil {
		containerData.Status = containerInfo.State.Status
		containerData.State = containerInfo.State.Status
	}

	if containerInfo.NetworkSettings != nil {
		containerData.IPAddress = containerInfo.NetworkSettings.IPAddress

		if containerInfo.NetworkSettings.Ports != nil {
			for port, bindings := range containerInfo.NetworkSettings.Ports {
				for _, binding := range bindings {
					containerData.Ports = append(containerData.Ports, container_types.Port{
						PrivatePort: int(port.Int()),
						PublicPort:  func() int { p, _ := strconv.Atoi(binding.HostPort); return p }(),
						Type:        port.Proto(),
					})
				}
			}
		}

		if containerInfo.NetworkSettings.Networks != nil {
			for name, network := range containerInfo.NetworkSettings.Networks {
				if network != nil {
					containerData.Networks = append(containerData.Networks, container_types.Network{
						Name:       name,
						IPAddress:  network.IPAddress,
						Gateway:    network.Gateway,
						MacAddress: network.MacAddress,
						Aliases:    network.Aliases,
					})
				}
			}
		}
	}

	if containerInfo.HostConfig != nil {
		containerData.HostConfig = container_types.HostConfig{
			Memory:     containerInfo.HostConfig.Memory,
			MemorySwap: containerInfo.HostConfig.MemorySwap,
			CPUShares:  containerInfo.HostConfig.CPUShares,
		}
	}

	for _, mount := range containerInfo.Mounts {
		containerData.Mounts = append(containerData.Mounts, container_types.Mount{
			Type:        string(mount.Type),
			Source:      mount.Source,
			Destination: mount.Destination,
			Mode:        mount.Mode,
		})
	}

	return containerData, nil
}
