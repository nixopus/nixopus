// Package docker exposes Docker operations from the internal deploy package
// for use by other modules like cloud.
package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	internaldocker "github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
)

// Repository is the interface for Docker operations.
// This is a re-export of the internal DockerRepository interface.
type Repository = internaldocker.DockerRepository

// Service is the Docker service implementation.
// This is a re-export of the internal DockerService.
type Service = internaldocker.DockerService

// GetDefaultService returns the default Docker service using the global DockerManager.
func GetDefaultService() (*Service, error) {
	manager := internaldocker.GetDockerManager()
	return manager.GetDefaultService()
}

// NewService creates a new Docker service.
func NewService() *Service {
	return internaldocker.NewDockerService()
}

// CreateContainerConfig holds the configuration for creating a dev container.
// This is a simplified config for the devrunner use case.
type CreateContainerConfig struct {
	Name         string
	Image        string
	Cmd          []string
	Env          []string
	WorkingDir   string
	ExposedPorts map[string]struct{}
	PortBindings map[string][]PortBinding
	Mounts       []Mount
	Labels       map[string]string
	Memory       int64
	CPUNanoCores int64
}

// PortBinding represents a port binding configuration
type PortBinding struct {
	HostIP   string
	HostPort string
}

// Mount represents a volume mount configuration
type Mount struct {
	Type     string
	Source   string
	Target   string
	ReadOnly bool
}

// CreateDevContainer creates a container configured for dev mode with volume mounts.
// This is a helper for the devrunner that wraps the lower-level Docker API.
func CreateDevContainer(ctx context.Context, svc *Service, cfg CreateContainerConfig) (string, error) {
	containerConfig := container.Config{
		Image:      cfg.Image,
		Cmd:        cfg.Cmd,
		WorkingDir: cfg.WorkingDir,
		Env:        cfg.Env,
		Labels:     cfg.Labels,
	}

	hostConfig := container.HostConfig{}

	// Set resource limits if specified
	if cfg.Memory > 0 || cfg.CPUNanoCores > 0 {
		hostConfig.Resources = container.Resources{
			Memory:   cfg.Memory,
			NanoCPUs: cfg.CPUNanoCores,
		}
	}

	resp, err := svc.CreateContainer(containerConfig, hostConfig, network.NetworkingConfig{}, cfg.Name)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

// StreamContainerLogs streams logs from a container
func StreamContainerLogs(ctx context.Context, svc *Service, containerID string, follow bool) (io.ReadCloser, error) {
	return svc.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     follow,
		Timestamps: true,
	})
}
