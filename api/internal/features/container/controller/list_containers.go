package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/container/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *ContainerController) ListContainers(f fuego.ContextNoBody) (*shared_types.Response, error) {
	containers, err := c.dockerService.ListAllContainers()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	var result []types.Container
	for _, container := range containers {
		containerInfo, err := c.dockerService.GetContainerById(container.ID)
		if err != nil {
			c.logger.Log(logger.Error, "Error inspecting container", container.ID)
			continue
		}

		containerData := types.Container{
			ID:        container.ID,
			Name:      "",
			Image:     container.Image,
			Status:    container.Status,
			State:     container.State,
			Created:   containerInfo.Created,
			Labels:    container.Labels,
			Command:   "",
			IPAddress: containerInfo.NetworkSettings.IPAddress,
			HostConfig: types.HostConfig{
				Memory:     containerInfo.HostConfig.Memory,
				MemorySwap: containerInfo.HostConfig.MemorySwap,
				CPUShares:  containerInfo.HostConfig.CPUShares,
			},
		}

		if container.Names != nil && len(container.Names) > 0 && len(container.Names[0]) > 1 {
			containerData.Name = container.Names[0][1:]
		}

		if containerInfo.Config != nil && containerInfo.Config.Cmd != nil && len(containerInfo.Config.Cmd) > 0 {
			containerData.Command = containerInfo.Config.Cmd[0]
		}

		for _, port := range container.Ports {
			containerData.Ports = append(containerData.Ports, types.Port{
				PrivatePort: int(port.PrivatePort),
				PublicPort:  int(port.PublicPort),
				Type:        port.Type,
			})
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

		result = append(result, containerData)
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Containers fetched successfully",
		Data:    result,
	}, nil
}
