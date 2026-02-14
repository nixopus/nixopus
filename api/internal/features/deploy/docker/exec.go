package docker

import (
	"bytes"
	"context"
	"fmt"
	"io"

	docker_types "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/google/uuid"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// GetRunningTaskContainerID finds the container ID for the running task of a Swarm service.
// Returns the container ID of the first running task, or an error if none found.
func (s *DockerService) GetRunningTaskContainerID(serviceID string) (string, error) {
	tasks, err := s.Cli.TaskList(s.Ctx, docker_types.TaskListOptions{
		Filters: filters.NewArgs(
			filters.Arg("service", serviceID),
			filters.Arg("desired-state", "running"),
		),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list tasks for service %s: %w", serviceID, err)
	}

	for _, task := range tasks {
		if task.Status.State == swarm.TaskStateRunning && task.Status.ContainerStatus != nil {
			containerID := task.Status.ContainerStatus.ContainerID
			if containerID != "" {
				return containerID, nil
			}
		}
	}

	return "", fmt.Errorf("no running container found for service %s", serviceID)
}

// ExecInContainer runs a command inside a running container and pipes stdin data to it.
// This is used to inject files into containers without bind mounts.
func (s *DockerService) ExecInContainer(containerID string, cmd []string, stdin io.Reader) error {
	execResp, err := s.Cli.ContainerExecCreate(s.Ctx, containerID, container.ExecOptions{
		AttachStdin:  stdin != nil,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          cmd,
	})
	if err != nil {
		return fmt.Errorf("failed to create exec: %w", err)
	}

	resp, err := s.Cli.ContainerExecAttach(s.Ctx, execResp.ID, container.ExecAttachOptions{})
	if err != nil {
		return fmt.Errorf("failed to attach to exec: %w", err)
	}
	defer resp.Close()

	// Pipe stdin data if provided
	if stdin != nil {
		if _, err := io.Copy(resp.Conn, stdin); err != nil {
			return fmt.Errorf("failed to write stdin data: %w", err)
		}
		resp.CloseWrite()
	}

	// Read all output to ensure exec completes
	var output bytes.Buffer
	io.Copy(&output, resp.Reader)

	// Check exec exit code
	inspect, err := s.Cli.ContainerExecInspect(s.Ctx, execResp.ID)
	if err != nil {
		return fmt.Errorf("failed to inspect exec result: %w", err)
	}

	if inspect.ExitCode != 0 {
		return fmt.Errorf("exec exited with code %d: %s", inspect.ExitCode, output.String())
	}

	return nil
}

// GetDockerServiceDirect returns the underlying DockerService (not the interface).
// This is needed for exec operations which are not part of the DockerRepository interface.
func GetDockerServiceDirect(ctx context.Context) (*DockerService, error) {
	orgIDAny := ctx.Value(shared_types.OrganizationIDKey)
	if orgIDAny == nil {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	var orgID uuid.UUID
	switch v := orgIDAny.(type) {
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("invalid organization ID: %w", err)
		}
		orgID = parsed
	case uuid.UUID:
		orgID = v
	default:
		return nil, fmt.Errorf("unexpected organization ID type: %T", v)
	}

	// Check cache for existing service
	if cached, ok := dockerServiceCache.Load(orgID); ok {
		entry := cached.(*cachedDockerService)
		if entry.service != nil && entry.service.Cli != nil {
			if _, err := entry.service.Cli.Ping(context.Background()); err == nil {
				return entry.service, nil
			}
		}
	}

	// Create new one if not cached
	svc, err := GetDockerServiceForOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Type assert - GetDockerServiceForOrganization returns *DockerService
	return svc, nil
}
