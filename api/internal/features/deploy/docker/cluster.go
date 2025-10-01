package docker

import (
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"
	"github.com/raghavyuva/nixopus-api/internal/config"
)

func (s *DockerService) InitCluster() error {
	config := config.AppConfig

	// Use localhost as default advertise address if SSH host is not configured
	// Useful during development
	advertiseAddr := "127.0.0.1:2377"
	if config.SSH.Host != "" {
		advertiseAddr = config.SSH.Host + ":2377"
	}

	_, err := s.Cli.SwarmInit(s.Ctx, swarm.InitRequest{
		ListenAddr: "0.0.0.0:2377",
		// Address that Hosts can use to reach the master node
		AdvertiseAddr: advertiseAddr,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *DockerService) JoinCluster() error {
	config := config.AppConfig

	// Use localhost as default advertise address if SSH host is not configured
	advertiseAddr := "127.0.0.1:2377"
	if config.SSH.Host != "" {
		advertiseAddr = config.SSH.Host + ":2377"
	}

	err := s.Cli.SwarmJoin(s.Ctx, swarm.JoinRequest{
		ListenAddr:    "0.0.0.0:2377",
		AdvertiseAddr: advertiseAddr,
	})
	return err
}

func (s *DockerService) LeaveCluster(force bool) error {
	err := s.Cli.SwarmLeave(s.Ctx, force)
	return err
}

func (s *DockerService) GetClusterInfo() (swarm.ClusterInfo, error) {
	clusterInfo, err := s.Cli.SwarmInspect(s.Ctx)
	return clusterInfo.ClusterInfo, err
}

func (s *DockerService) GetClusterNodes() ([]swarm.Node, error) {
	nodes, err := s.Cli.NodeList(s.Ctx, types.NodeListOptions{})
	return nodes, err
}

func (s *DockerService) GetClusterServices() ([]swarm.Service, error) {
	services, err := s.Cli.ServiceList(s.Ctx, types.ServiceListOptions{})
	return services, err
}

func (s *DockerService) GetClusterTasks() ([]swarm.Task, error) {
	tasks, err := s.Cli.TaskList(s.Ctx, types.TaskListOptions{})
	return tasks, err
}

func (s *DockerService) GetClusterSecrets() ([]swarm.Secret, error) {
	secrets, err := s.Cli.SecretList(s.Ctx, types.SecretListOptions{})
	return secrets, err
}

func (s *DockerService) GetClusterConfigs() ([]swarm.Config, error) {
	configs, err := s.Cli.ConfigList(s.Ctx, types.ConfigListOptions{})
	return configs, err
}

func (s *DockerService) GetClusterVolumes() ([]*volume.Volume, error) {
	volumes, err := s.Cli.VolumeList(s.Ctx, volume.ListOptions{})
	return volumes.Volumes, err
}

func (s *DockerService) GetClusterNetworks() ([]network.Summary, error) {
	networks, err := s.Cli.NetworkList(s.Ctx, network.ListOptions{})
	return networks, err
}

func (s *DockerService) UpdateNodeAvailability(nodeID string, availability swarm.NodeAvailability) error {
	node, _, err := s.Cli.NodeInspectWithRaw(s.Ctx, nodeID)
	if err != nil {
		return err
	}
	spec := node.Spec
	spec.Availability = availability
	return s.Cli.NodeUpdate(s.Ctx, nodeID, node.Version, spec)
}

func (s *DockerService) ListenEvents(opts events.ListOptions) (<-chan events.Message, <-chan error) {
	return s.Cli.Events(s.Ctx, opts)
}

func (s *DockerService) ScaleService(serviceID string, replicas uint64, rollback string) error {
	svc, _, err := s.Cli.ServiceInspectWithRaw(s.Ctx, serviceID, types.ServiceInspectOptions{})
	if err != nil {
		return err
	}
	spec := svc.Spec
	spec.Mode.Replicated.Replicas = &replicas
	_, err = s.Cli.ServiceUpdate(s.Ctx, serviceID, svc.Version, spec, types.ServiceUpdateOptions{
		Rollback: rollback,
	})
	return err
}

func (s *DockerService) GetServiceHealth(service swarm.Service) (int, int, error) {
	tasks, err := s.Cli.TaskList(s.Ctx, types.TaskListOptions{
		Filters: filters.NewArgs(
			filters.Arg("service", service.ID),
		),
	})
	if err != nil {
		return 0, 0, err
	}

	running := 0
	for _, t := range tasks {
		if t.Status.State == swarm.TaskStateRunning {
			running++
		}
	}
	desired := 0
	if service.Spec.Mode.Replicated != nil && service.Spec.Mode.Replicated.Replicas != nil {
		desired = int(*service.Spec.Mode.Replicated.Replicas)
	}
	return running, desired, nil
}

func (s *DockerService) GetTaskHealth(task swarm.Task) swarm.TaskState {
	if task.Status.State != "" {
		return task.Status.State
	}
	return swarm.TaskState("")
}

func (s *DockerService) CreateService(service swarm.Service) error {
	_, err := s.Cli.ServiceCreate(s.Ctx, service.Spec, types.ServiceCreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (s *DockerService) UpdateService(serviceID string, serviceSpec swarm.ServiceSpec, rollback string) error {
	svc, _, err := s.Cli.ServiceInspectWithRaw(s.Ctx, serviceID, types.ServiceInspectOptions{})
	if err != nil {
		return err
	}

	_, err = s.Cli.ServiceUpdate(s.Ctx, serviceID, svc.Version, serviceSpec, types.ServiceUpdateOptions{
		Rollback: rollback,
	})
	return err
}

func (s *DockerService) DeleteService(serviceID string) error {
	err := s.Cli.ServiceRemove(s.Ctx, serviceID)
	return err
}

func (s *DockerService) RollbackService(serviceID string) error {
	_, err := s.Cli.ServiceUpdate(s.Ctx, serviceID, swarm.Version{}, swarm.ServiceSpec{}, types.ServiceUpdateOptions{
		Rollback: "previous",
	})
	return err
}

func (s *DockerService) GetServiceByID(serviceID string) (swarm.Service, error) {
	service, _, err := s.Cli.ServiceInspectWithRaw(s.Ctx, serviceID, types.ServiceInspectOptions{})
	return service, err
}
