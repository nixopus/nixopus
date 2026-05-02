package service

import (
	"context"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/nixopus/nixopus/api/internal/features/deploy/docker"
)

// stubDockerRepository implements docker.DockerRepository for unit tests.
// Only non-nil hooks are consulted; omitted behavior returns zeros.
type stubDockerRepository struct {
	listContainers    func(opts container.ListOptions) ([]container.Summary, error)
	listAllContainers func() ([]container.Summary, error)
	listAllImages     func(opts image.ListOptions) []image.Summary
	getContainerByID  func(id string) (container.InspectResponse, error)
	getContainerLogs  func(containerID string, opts container.LogsOptions) (io.Reader, error)
	startContainer    func(containerID string, opts container.StartOptions) error
	stopContainer     func(containerID string, opts container.StopOptions) error
	removeContainer   func(containerID string, opts container.RemoveOptions) error
	restartContainer  func(containerID string, opts container.StopOptions) error
	updateResources   func(containerID string, cfg container.UpdateConfig) (container.ContainerUpdateOKBody, error)
	pruneBuildCache   func(opts types.BuildCachePruneOptions) error
	pruneImages       func(args filters.Args) (image.PruneReport, error)
}

var _ docker.DockerRepository = (*stubDockerRepository)(nil)

func (s *stubDockerRepository) ListAllContainers() ([]container.Summary, error) {
	if s.listAllContainers != nil {
		return s.listAllContainers()
	}
	return nil, nil
}

func (s *stubDockerRepository) ListContainers(opts container.ListOptions) ([]container.Summary, error) {
	if s.listContainers != nil {
		return s.listContainers(opts)
	}
	return nil, nil
}

func (s *stubDockerRepository) ListAllImages(opts image.ListOptions) []image.Summary {
	if s.listAllImages != nil {
		return s.listAllImages(opts)
	}
	return nil
}

func (s *stubDockerRepository) StopContainer(containerID string, opts container.StopOptions) error {
	if s.stopContainer != nil {
		return s.stopContainer(containerID, opts)
	}
	return nil
}

func (s *stubDockerRepository) RemoveContainer(containerID string, opts container.RemoveOptions) error {
	if s.removeContainer != nil {
		return s.removeContainer(containerID, opts)
	}
	return nil
}

func (s *stubDockerRepository) StartContainer(containerID string, opts container.StartOptions) error {
	if s.startContainer != nil {
		return s.startContainer(containerID, opts)
	}
	return nil
}

func (s *stubDockerRepository) GetContainerLogs(containerID string, opts container.LogsOptions) (io.Reader, error) {
	if s.getContainerLogs != nil {
		return s.getContainerLogs(containerID, opts)
	}
	return nil, nil
}

func (s *stubDockerRepository) GetContainerById(containerID string) (container.InspectResponse, error) {
	if s.getContainerByID != nil {
		return s.getContainerByID(containerID)
	}
	return container.InspectResponse{}, nil
}

func (s *stubDockerRepository) GetImageById(imageID string, opts client.ImageInspectOption) (image.InspectResponse, error) {
	return image.InspectResponse{}, nil
}

func (s *stubDockerRepository) ImagePull(ctx context.Context, ref string, opts image.PullOptions) (io.ReadCloser, error) {
	return nil, nil
}

func (s *stubDockerRepository) BuildImage(opts types.ImageBuildOptions, buildContext io.Reader) (types.ImageBuildResponse, error) {
	return types.ImageBuildResponse{}, nil
}

func (s *stubDockerRepository) CreateContainer(cfg container.Config, hostConfig container.HostConfig, networkConfig network.NetworkingConfig, containerName string) (container.CreateResponse, error) {
	return container.CreateResponse{}, nil
}

func (s *stubDockerRepository) ContainerLogs(ctx context.Context, containerID string, opts container.LogsOptions) (io.ReadCloser, error) {
	return nil, nil
}

func (s *stubDockerRepository) RestartContainer(containerID string, opts container.StopOptions) error {
	if s.restartContainer != nil {
		return s.restartContainer(containerID, opts)
	}
	return nil
}

func (s *stubDockerRepository) UpdateContainerResources(containerID string, resources container.UpdateConfig) (container.ContainerUpdateOKBody, error) {
	if s.updateResources != nil {
		return s.updateResources(containerID, resources)
	}
	return container.ContainerUpdateOKBody{}, nil
}

func (s *stubDockerRepository) ComposeUp(composeFilePath string, envVars map[string]string, overrideFiles ...string) (string, error) {
	return "", nil
}

func (s *stubDockerRepository) ComposeUpWithCallback(composeFilePath string, envVars map[string]string, outputCallback func(string), overrideFiles ...string) (string, error) {
	return "", nil
}

func (s *stubDockerRepository) ComposeDown(composeFilePath string, envVars map[string]string, overrideFiles ...string) error {
	return nil
}

func (s *stubDockerRepository) ComposeDownWithCallback(composeFilePath string, envVars map[string]string, outputCallback func(string), overrideFiles ...string) error {
	return nil
}

func (s *stubDockerRepository) ComposeRestart(composeFilePath string, envVars map[string]string, outputCallback func(string), overrideFiles ...string) error {
	return nil
}

func (s *stubDockerRepository) ComposeBuild(composeFilePath string, envVars map[string]string, overrideFiles ...string) error {
	return nil
}

func (s *stubDockerRepository) RemoveImage(imageName string, opts image.RemoveOptions) error {
	return nil
}

func (s *stubDockerRepository) PruneBuildCache(opts types.BuildCachePruneOptions) error {
	if s.pruneBuildCache != nil {
		return s.pruneBuildCache(opts)
	}
	return nil
}

func (s *stubDockerRepository) PruneImages(opts filters.Args) (image.PruneReport, error) {
	if s.pruneImages != nil {
		return s.pruneImages(opts)
	}
	return image.PruneReport{}, nil
}

func (s *stubDockerRepository) InitCluster() error {
	return nil
}

func (s *stubDockerRepository) JoinCluster() error {
	return nil
}

func (s *stubDockerRepository) LeaveCluster(force bool) error {
	return nil
}

func (s *stubDockerRepository) GetClusterInfo() (swarm.ClusterInfo, error) {
	return swarm.ClusterInfo{}, nil
}

func (s *stubDockerRepository) GetClusterNodes() ([]swarm.Node, error) {
	return nil, nil
}

func (s *stubDockerRepository) GetClusterServices() ([]swarm.Service, error) {
	return nil, nil
}

func (s *stubDockerRepository) GetClusterTasks() ([]swarm.Task, error) {
	return nil, nil
}

func (s *stubDockerRepository) GetTasksByServiceID(serviceID string) ([]swarm.Task, error) {
	return nil, nil
}

func (s *stubDockerRepository) GetClusterSecrets() ([]swarm.Secret, error) {
	return nil, nil
}

func (s *stubDockerRepository) GetClusterConfigs() ([]swarm.Config, error) {
	return nil, nil
}

func (s *stubDockerRepository) GetClusterVolumes() ([]*volume.Volume, error) {
	return nil, nil
}

func (s *stubDockerRepository) GetClusterNetworks() ([]network.Summary, error) {
	return nil, nil
}

func (s *stubDockerRepository) UpdateNodeAvailability(nodeID string, availability swarm.NodeAvailability) error {
	return nil
}

func (s *stubDockerRepository) ScaleService(serviceID string, replicas uint64, rollback string) error {
	return nil
}

func (s *stubDockerRepository) ListenEvents(opts events.ListOptions) (<-chan events.Message, <-chan error) {
	ch := make(chan events.Message)
	ech := make(chan error)
	close(ch)
	close(ech)
	return ch, ech
}

func (s *stubDockerRepository) GetServiceHealth(service swarm.Service) (int, int, error) {
	return 0, 0, nil
}

func (s *stubDockerRepository) GetTaskHealth(task swarm.Task) swarm.TaskState {
	return swarm.TaskState("")
}

func (s *stubDockerRepository) CreateService(service swarm.Service) error {
	return nil
}

func (s *stubDockerRepository) UpdateService(serviceID string, serviceSpec swarm.ServiceSpec, rollback string) error {
	return nil
}

func (s *stubDockerRepository) DeleteService(serviceID string) error {
	return nil
}

func (s *stubDockerRepository) RollbackService(serviceID string) error {
	return nil
}

func (s *stubDockerRepository) GetServiceByID(serviceID string) (swarm.Service, error) {
	return swarm.Service{}, nil
}

func (s *stubDockerRepository) GetServiceByName(name string) (*swarm.Service, error) {
	return nil, nil
}

func (s *stubDockerRepository) GetServiceByLabel(key, value string) (*swarm.Service, error) {
	return nil, nil
}
