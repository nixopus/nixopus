package controller

import (
	"net/http"
	"strconv"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *ContainerController) GetContainer(f fuego.ContextNoBody) (*shared_types.Response, error) {
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
		ID:        containerInfo.ID,
		Name:      containerInfo.Name[1:],
		Image:     containerInfo.Config.Image,
		Status:    containerInfo.State.Status,
		State:     containerInfo.State.Status,
		Created:   containerInfo.Created,
		Labels:    containerInfo.Config.Labels,
		Command:   containerInfo.Config.Cmd[0],
		IPAddress: containerInfo.NetworkSettings.IPAddress,
		HostConfig: types.HostConfig{
			Memory:     containerInfo.HostConfig.Memory,
			MemorySwap: containerInfo.HostConfig.MemorySwap,
			CPUShares:  containerInfo.HostConfig.CPUShares,
		},
	}

	for port, bindings := range containerInfo.NetworkSettings.Ports {
		for _, binding := range bindings {
			containerData.Ports = append(containerData.Ports, types.Port{
				PrivatePort: int(port.Int()),
				PublicPort:  func() int { p, _ := strconv.Atoi(binding.HostPort); return p }(),
				Type:        port.Proto(),
			})
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

	for name, network := range containerInfo.NetworkSettings.Networks {
		containerData.Networks = append(containerData.Networks, types.Network{
			Name:       name,
			IPAddress:  network.IPAddress,
			Gateway:    network.Gateway,
			MacAddress: network.MacAddress,
			Aliases:    network.Aliases,
		})
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Container fetched successfully",
		Data:    containerData,
	}, nil
}
