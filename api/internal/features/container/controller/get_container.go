package controller

import (
	"net/http"
	"strconv"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

func (c *ContainerController) GetContainer(f fuego.ContextNoBody) (*types.GetContainerResponse, error) {
	containerID := f.PathParam("container_id")

	containerInfo, err := c.dockerService.GetContainerById(containerID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	containerData := types.Container{
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
					containerData.Ports = append(containerData.Ports, types.Port{
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
					containerData.Networks = append(containerData.Networks, types.Network{
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
		containerData.HostConfig = types.HostConfig{
			Memory:     containerInfo.HostConfig.Memory,
			MemorySwap: containerInfo.HostConfig.MemorySwap,
			CPUShares:  containerInfo.HostConfig.CPUShares,
		}
	}

	for _, mount := range containerInfo.Mounts {
		containerData.Mounts = append(containerData.Mounts, types.Mount{
			Type:        string(mount.Type),
			Source:      mount.Source,
			Destination: mount.Destination,
			Mode:        mount.Mode,
		})
	}

	return &types.GetContainerResponse{
		Status:  "success",
		Message: "Container fetched successfully",
		Data:    containerData,
	}, nil
}
