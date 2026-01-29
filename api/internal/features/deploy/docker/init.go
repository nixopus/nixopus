package docker

import (
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
)

type DockerService struct {
	Cli    *client.Client
	Ctx    context.Context
	logger logger.Logger
}

type DockerRepository interface {
	ListAllContainers() ([]container.Summary, error)
	ListContainers(opts container.ListOptions) ([]container.Summary, error)
	ListAllImages(opts image.ListOptions) []image.Summary

	StopContainer(containerID string, opts container.StopOptions) error
	RemoveContainer(containerID string, opts container.RemoveOptions) error
	StartContainer(containerID string, opts container.StartOptions) error
	GetContainerLogs(containerID string, opts container.LogsOptions) (io.Reader, error)
	GetContainerById(containerID string) (container.InspectResponse, error)
	GetImageById(imageID string, opts client.ImageInspectOption) (image.InspectResponse, error)

	BuildImage(opts types.ImageBuildOptions, buildContext io.Reader) (types.ImageBuildResponse, error)
	CreateContainer(config container.Config, hostConfig container.HostConfig, networkConfig network.NetworkingConfig, containerName string) (container.CreateResponse, error)
	// CreateDeployment(deployment *deploy_types.CreateDeploymentRequest, userID uuid.UUID, contextPath string) error
	ContainerLogs(ctx context.Context, containerID string, opts container.LogsOptions) (io.ReadCloser, error)
	RestartContainer(containerID string, opts container.StopOptions) error

	ComposeUp(composeFilePath string, envVars map[string]string) error
	ComposeDown(composeFilePath string) error
	ComposeBuild(composeFilePath string, envVars map[string]string) error
	RemoveImage(imageName string, opts image.RemoveOptions) error
	PruneBuildCache(opts types.BuildCachePruneOptions) error
	PruneImages(opts filters.Args) (image.PruneReport, error)

	InitCluster() error
	JoinCluster() error
	LeaveCluster(force bool) error
	GetClusterInfo() (swarm.ClusterInfo, error)
	GetClusterNodes() ([]swarm.Node, error)
	GetClusterServices() ([]swarm.Service, error)
	GetClusterTasks() ([]swarm.Task, error)
	GetClusterSecrets() ([]swarm.Secret, error)
	GetClusterConfigs() ([]swarm.Config, error)
	GetClusterVolumes() ([]*volume.Volume, error)
	GetClusterNetworks() ([]network.Summary, error)
	UpdateNodeAvailability(nodeID string, availability swarm.NodeAvailability) error
	ScaleService(serviceID string, replicas uint64, rollback string) error
	ListenEvents(opts events.ListOptions) (<-chan events.Message, <-chan error)
	GetServiceHealth(service swarm.Service) (int, int, error)
	GetTaskHealth(task swarm.Task) swarm.TaskState
	CreateService(service swarm.Service) error
	UpdateService(serviceID string, serviceSpec swarm.ServiceSpec, rollback string) error
	DeleteService(serviceID string) error
	RollbackService(serviceID string) error
	GetServiceByID(serviceID string) (swarm.Service, error)
}

type DockerClient struct {
	Client *client.Client
}

// DockerClientConfig represents configuration for a Docker client connection
type DockerClientConfig struct {
	Host    string // Docker daemon host (e.g., "unix:///var/run/docker.sock", "tcp://host:2376")
	Context context.Context
	Logger  logger.Logger
}

// DockerManager manages multiple Docker client connections to different Docker daemons/sockets
// For now, it defaults to single client mode for backward compatibility
// In the future, it can be extended to support multiple Docker hosts/daemons
type DockerManager struct {
	clients   map[string]*client.Client      // Map of client ID to Docker client
	configs   map[string]*DockerClientConfig // Map of client ID to config
	defaultID string                         // ID of the default client
	mu        sync.RWMutex                   // Mutex for thread safe access
}

var (
	// globalDockerManager is the singleton instance of DockerManager
	globalDockerManager *DockerManager
	globalDockerMu      sync.Once
)

// GetDockerManager returns the global singleton DockerManager instance
// This ensures we have a single DockerManager instance across the entire application
// It's initialized lazily on first access with the default Docker client
//
// IMPORTANT: Cluster initialization is performed automatically during the first call
// to GetDockerManager(). This lazy initialization may introduce unpredictable latency
// if the first access happens during a request-handling path. The cluster initialization
// logic (lines 130-143) executes synchronously and may take several seconds to complete.
//
// For production deployments, consider:
//   - Performing cluster initialization explicitly during application startup
//   - Pre-warming the DockerManager by calling GetDockerManager() during initialization
//   - Monitoring the first request latency to identify any initialization delays
//
// Cluster initialization is performed automatically (same as NewDockerService)
func GetDockerManager() *DockerManager {
	globalDockerMu.Do(func() {
		defaultClient, actualHost := NewDockerClient()
		defaultLogger := logger.NewLogger()
		defaultCtx := context.Background()

		globalDockerManager = &DockerManager{
			clients:   make(map[string]*client.Client),
			configs:   make(map[string]*DockerClientConfig),
			defaultID: "default",
		}
		globalDockerManager.clients["default"] = defaultClient
		globalDockerManager.configs["default"] = &DockerClientConfig{
			Host:    actualHost, // Use the actual host that was connected to
			Context: defaultCtx,
			Logger:  defaultLogger,
		}

		// Initialize cluster if not already initialized (same behavior as NewDockerService)
		// This should be run on master node only
		// TODO: Add a check to see if the node is the master node
		// WARNING: This should be thought again during multi-server architecture feature
		if !isClusterInitialized(defaultClient) {
			tempService := &DockerService{
				Cli:    defaultClient,
				Ctx:    defaultCtx,
				logger: defaultLogger,
			}
			if err := tempService.InitCluster(); err != nil {
				defaultLogger.Log(logger.Warning, "Failed to initialize cluster", err.Error())
			} else {
				defaultLogger.Log(logger.Info, "Cluster initialized successfully", "")
			}
		} else {
			defaultLogger.Log(logger.Info, "Cluster already initialized", "")
		}
	})
	return globalDockerManager
}

// AddClient adds a new Docker client connection to the manager with a unique ID
// The host can be a Unix socket (unix:///path/to/socket) or TCP (tcp://host:port)
// Example: AddClient("server1", "tcp://192.168.1.100:2376")
// Example: AddClient("server2", "unix:///var/run/docker.sock")
func (m *DockerManager) AddClient(id string, host string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "" {
		return fmt.Errorf("client ID cannot be empty")
	}
	if host == "" {
		return fmt.Errorf("Docker host cannot be empty")
	}

	// Check if an existing client exists and close it before overwriting
	if existingClient, exists := m.clients[id]; exists {
		if err := existingClient.Close(); err != nil {
			// Log the error but don't fail - we'll still replace the client
			if config, ok := m.configs[id]; ok {
				config.Logger.Log(logger.Warning, fmt.Sprintf("Failed to close existing Docker client %s", id), err.Error())
			} else {
				// Fallback to default logger if config doesn't exist
				defaultLogger := logger.NewLogger()
				defaultLogger.Log(logger.Warning, fmt.Sprintf("Failed to close existing Docker client %s", id), err.Error())
			}
		}
	}

	cli, err := client.NewClientWithOpts(
		client.WithHost(host),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		// Try with TLS if available
		cli, err = client.NewClientWithOpts(
			client.WithHost(host),
			client.WithAPIVersionNegotiation(),
			client.WithTLSClientConfigFromEnv(),
		)
		if err != nil {
			return fmt.Errorf("failed to create Docker client for %s: %w", host, err)
		}
	}

	m.clients[id] = cli
	m.configs[id] = &DockerClientConfig{
		Host:    host,
		Context: context.Background(),
		Logger:  logger.NewLogger(),
	}
	return nil
}

// GetClient retrieves a Docker client by ID, or returns the default client if ID is empty
func (m *DockerManager) GetClient(id string) (*client.Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if id == "" {
		id = m.defaultID
	}

	client, exists := m.clients[id]
	if !exists {
		return nil, fmt.Errorf("Docker client with ID '%s' not found", id)
	}

	return client, nil
}

// GetService retrieves a DockerService for a specific client ID
// This wraps the client with the service interface for backward compatibility
func (m *DockerManager) GetService(id string) (*DockerService, error) {
	cli, err := m.GetClient(id)
	if err != nil {
		return nil, err
	}

	m.mu.RLock()
	config, exists := m.configs[id]
	m.mu.RUnlock()

	if !exists {
		config = &DockerClientConfig{
			Host:    "unix:///var/run/docker.sock",
			Context: context.Background(),
			Logger:  logger.NewLogger(),
		}
	}

	return &DockerService{
		Cli:    cli,
		Ctx:    config.Context,
		logger: config.Logger,
	}, nil
}

// SetDefault sets the default client ID
func (m *DockerManager) SetDefault(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[id]; !exists {
		return fmt.Errorf("Docker client with ID '%s' does not exist", id)
	}

	m.defaultID = id
	return nil
}

// ListClients returns a list of all client IDs
func (m *DockerManager) ListClients() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.clients))
	for id := range m.clients {
		ids = append(ids, id)
	}
	return ids
}

// GetDefaultService returns the default Docker service
func (m *DockerManager) GetDefaultService() (*DockerService, error) {
	return m.GetService("")
}

// NewDockerService creates a new instance of DockerService using the default docker client.
func NewDockerService() *DockerService {
	client, _ := NewDockerClient()
	service := &DockerService{
		Cli:    client,
		Ctx:    context.Background(),
		logger: logger.NewLogger(),
	}

	// Initialize cluster if not already initialized, this should be run on master node only
	// TODO: Add a check to see if the node is the master node
	// WARNING: This should be thought again during multi-server architecture feature
	if !isClusterInitialized(client) {
		if err := service.InitCluster(); err != nil {
			service.logger.Log(logger.Warning, "Failed to initialize cluster", err.Error())
		} else {
			service.logger.Log(logger.Info, "Cluster initialized successfully", "")
		}
	} else {
		service.logger.Log(logger.Info, "Cluster already initialized", "")
	}

	return service
}

func isClusterInitialized(cli *client.Client) bool {
	info, err := cli.Info(context.Background())
	if err != nil {
		return false
	}
	return info.Swarm.LocalNodeState == swarm.LocalNodeStateActive
}

func NewDockerServiceWithClient(cli *client.Client, ctx context.Context, logger logger.Logger) *DockerService {
	return &DockerService{
		Cli:    cli,
		Ctx:    ctx,
		logger: logger,
	}
}

// NewDockerClient creates a new docker client with the environment variables and
// the correct API version negotiation.
// Returns the client and the actual host string that was used for the connection.
func NewDockerClient() (*client.Client, string) {
	defaultHost := "unix:///var/run/docker.sock"
	cli, err := client.NewClientWithOpts(
		client.WithHost(defaultHost),
		client.WithAPIVersionNegotiation(),
	)
	if err == nil {
		return cli, defaultHost
	}

	// Try with FromEnv and TLS
	cli, err = client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
		client.WithTLSClientConfigFromEnv(),
	)
	if err == nil {
		// Extract host from environment or use default
		host := getHostFromEnv()
		return cli, host
	}

	// Try with FromEnv without TLS
	cli, err = client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)
	if err == nil {
		host := getHostFromEnv()
		return cli, host
	}

	panic(err)
}

// getHostFromEnv extracts the Docker host from environment variables
// Returns the host from DOCKER_HOST env var, or default socket if not set
func getHostFromEnv() string {
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		return host
	}
	return "unix:///var/run/docker.sock"
}

// ListAllContainers returns a list of all containers running on the host, along with their
// IDs, names, and statuses. The returned list is sorted by container ID in ascending order.
//
// If an error occurs while listing the containers, it returns the error (no panic).
func (s *DockerService) ListAllContainers() ([]container.Summary, error) {
	containers, err := s.Cli.ContainerList(s.Ctx, container.ListOptions{
		All: true,
	})
	if err != nil {
		return nil, err
	}

	return containers, nil
}

// ListContainers returns containers using the provided docker list options
// (including native filters like name/status/ancestor and optional limits).
func (s *DockerService) ListContainers(opts container.ListOptions) ([]container.Summary, error) {
	containers, err := s.Cli.ContainerList(s.Ctx, opts)
	if err != nil {
		return nil, err
	}
	return containers, nil
}

// StopContainer stops the container with the given ID. If the container does not exist,
// it returns a docker client error. If any other error occurs while stopping the container,
// it panics with the error.
func (s *DockerService) StopContainer(containerID string, opts container.StopOptions) error {
	return s.Cli.ContainerStop(s.Ctx, containerID, opts)
}

// RemoveContainer removes the container with the given ID. If the container does not exist,
// it returns a docker client error. If any other error occurs while removing the container,
// it panics with the error.
func (s *DockerService) RemoveContainer(containerID string, opts container.RemoveOptions) error {
	return s.Cli.ContainerRemove(s.Ctx, containerID, opts)
}

// StartContainer starts a container with the given ID. If the container does not exist,
// it returns a docker client error. If any other error occurs while starting the container,
// it panics with the error.
func (s *DockerService) StartContainer(containerID string, opts container.StartOptions) error {
	return s.Cli.ContainerStart(s.Ctx, containerID, opts)
}

func (s *DockerService) RestartContainer(containerID string, opts container.StopOptions) error {
	return s.Cli.ContainerRestart(s.Ctx, containerID, opts)
}

// GetContainerLogs retrieves the logs of a container with the given ID, using the given opts.
//
// The returned io.Reader is a stream of the container's logs. If the container does not exist,
// it returns a docker client error. If any other error occurs while retrieving the logs,
// it panics with the error.
func (s *DockerService) GetContainerLogs(containerID string, opts container.LogsOptions) (io.Reader, error) {
	var logs io.Reader
	logs, err := s.Cli.ContainerLogs(s.Ctx, containerID, opts)
	if err != nil {
		return nil, err
	}

	return logs, nil
}

// GetContainerById retrieves the detailed information of a container by its ID.
//
// It queries the Docker client to inspect the container associated with the
// provided containerID. If successful, it returns the container's inspection
// data. If the container does not exist or any error occurs during the
// inspection, it returns an error.
//
// Parameters:
//
//	containerID - the unique identifier of the container to inspect.
//
// Returns:
//
//	container.InspectResponse - the detailed information of the container if found.
//	error - an error if the inspection fails or the container does not exist.
func (s *DockerService) GetContainerById(containerID string) (container.InspectResponse, error) {
	return s.Cli.ContainerInspect(s.Ctx, containerID)
}

// ListAllImages returns a list of all images on the host, using the given opts.
//
// The returned list is a slice of image.Summary structs, which contain the
// image ID, repository name, tags, and size of each image on the host. If
// an error occurs while listing the images, it returns an empty slice.
func (s *DockerService) ListAllImages(opts image.ListOptions) []image.Summary {
	images, err := s.Cli.ImageList(s.Ctx, opts)
	if err != nil {
		s.logger.Log(logger.Error, "Failed to list images", err.Error())
		return []image.Summary{}
	}
	return images
}

// GetImageById retrieves the detailed information of an image by its ID.
//
// It queries the Docker client to inspect the image associated with the
// provided imageID. If successful, it returns the image's inspection data.
// If the image does not exist or any error occurs during the inspection, it
// returns an error.
//
// Parameters:
//
//	imageID - the unique identifier of the image to inspect.
//	opts - an optional set of options for the inspection.
//
// Returns:
//
//	image.InspectResponse - the detailed information of the image if found.
//	error - an error if the inspection fails or the image does not exist.
func (s *DockerService) GetImageById(imageID string, opts client.ImageInspectOption) (image.InspectResponse, error) {
	return s.Cli.ImageInspect(s.Ctx, imageID, opts)
}

// BuildImage builds a Docker image using the specified build options.
//
// This function uses the Docker client to build a Docker image based on the
// provided options. It returns an ImageBuildResponse that contains the
// details of the build process, such as the image ID and build logs. If an
// error occurs during the build process, it returns an error.
//
// Parameters:
//
//	opts - the ImageBuildOptions that specify the build context, Dockerfile,
//	and other build configurations.
//
// Returns:
//
//	types.ImageBuildResponse - the response containing the build details.
//	error - an error if the build fails.
func (s *DockerService) BuildImage(opts types.ImageBuildOptions, buildContext io.Reader) (types.ImageBuildResponse, error) {
	return s.Cli.ImageBuild(s.Ctx, buildContext, opts)
}

// CreateContainer creates a new Docker container with the specified configurations.
//
// It uses the provided container configuration, host configuration, and networking configuration
// to create the container. The container is created with the specified name and is set to run
// on the "amd64" architecture and "linux" OS platform.
//
// Parameters:
//
//	config - the configuration for the container, including environment variables, commands, etc.
//	hostConfig - the host-specific configuration for the container, such as resource limits.
//	networkConfig - the networking configuration for the container, specifying network settings.
//	containerName - the name to assign to the new container.
//
// Returns:
//
//	container.CreateResponse - the response containing details of the created container.
//	error - an error if the container creation fails.
func (s *DockerService) CreateContainer(config container.Config, hostConfig container.HostConfig, networkConfig network.NetworkingConfig, containerName string) (container.CreateResponse, error) {
	return s.Cli.ContainerCreate(s.Ctx, &config, &hostConfig, &networkConfig, &v1.Platform{}, containerName)
}

// ContainerLogs retrieves the logs of a container with the given ID, using the specified options.
//
// This function returns an io.ReadCloser that provides a stream of the container's logs. The logs can include
// both stdout and stderr depending on the provided options. If the container does not exist or any error occurs
// while retrieving the logs, an error is returned.
//
// Parameters:
//
//	Ctx - the context within which the logs are retrieved, allowing for cancellation and timeouts.
//	containerID - the unique identifier of the container whose logs are to be retrieved.
//	opts - the options specifying which logs to retrieve, including whether to include stdout, stderr, and timestamps.
//
// Returns:
//
//	io.ReadCloser - a stream of the container's logs.
//	error - an error if the container does not exist or if there is an issue retrieving the logs.
func (s *DockerService) ContainerLogs(Ctx context.Context, containerID string, opts container.LogsOptions) (io.ReadCloser, error) {
	return s.Cli.ContainerLogs(Ctx, containerID, opts)
}

// ComposeUp starts the Docker Compose services defined in the specified compose file
func (s *DockerService) ComposeUp(composeFilePath string, envVars map[string]string) error {
	manager, err := ssh.GetSSHManagerFromContext(s.Ctx)
	if err != nil {
		return fmt.Errorf("failed to get SSH manager: %w", err)
	}
	envVarsStr := ""
	for k, v := range envVars {
		envVarsStr += fmt.Sprintf("export %s=%s && ", k, v)
	}
	// Use --force-recreate to handle existing containers and --remove-orphans to clean up old containers
	command := fmt.Sprintf("%sdocker compose -f %s up -d --force-recreate --remove-orphans 2>&1", envVarsStr, composeFilePath)
	output, err := manager.RunCommand(command)
	if err != nil {
		return fmt.Errorf("failed to start docker compose services: %v, output: %s", err, output)
	}
	return nil
}

// ComposeDown stops and removes the Docker Compose services
func (s *DockerService) ComposeDown(composeFilePath string) error {
	manager, err := ssh.GetSSHManagerFromContext(s.Ctx)
	if err != nil {
		return fmt.Errorf("failed to get SSH manager: %w", err)
	}
	command := fmt.Sprintf("docker compose -f %s down", composeFilePath)
	output, err := manager.RunCommand(command)
	if err != nil {
		return fmt.Errorf("failed to stop docker compose services: %v, output: %s", err, output)
	}
	return nil
}

// ComposeBuild builds the Docker Compose services
func (s *DockerService) ComposeBuild(composeFilePath string, envVars map[string]string) error {
	manager, err := ssh.GetSSHManagerFromContext(s.Ctx)
	if err != nil {
		return fmt.Errorf("failed to get SSH manager: %w", err)
	}
	envVarsStr := ""
	for k, v := range envVars {
		envVarsStr += fmt.Sprintf("export %s=%s && ", k, v)
	}
	command := fmt.Sprintf("%sdocker compose -f %s build", envVarsStr, composeFilePath)
	output, err := manager.RunCommand(command)
	if err != nil {
		return fmt.Errorf("failed to build docker compose services: %v, output: %s", err, output)
	}
	return nil
}

// RemoveImage removes an image from the Docker host.
//
// This function removes an image from the Docker host using the Docker client.
// It takes an image name and an optional set of options for the removal process.
//
// Parameters:
//
//	imageName - the name of the image to remove.
//	opts - an optional set of options for the removal process.
//
// Returns:
//
//	error - an error if the image removal fails.
func (s *DockerService) RemoveImage(imageName string, opts image.RemoveOptions) error {
	ctx := context.Background()
	_, err := s.Cli.ImageRemove(ctx, imageName, image.RemoveOptions{
		Force:         opts.Force,
		PruneChildren: true,
	})
	return err
}

// PruneImages prunes the images on the Docker host.
//
// This function prunes the images on the Docker host using the Docker client.
// It takes an optional set of options for the pruning process.
//
// Parameters:

func (s *DockerService) PruneBuildCache(opts types.BuildCachePruneOptions) error {
	_, err := s.Cli.BuildCachePrune(s.Ctx, opts)
	return err
}

func (s *DockerService) PruneImages(opts filters.Args) (image.PruneReport, error) {
	pruneReport, err := s.Cli.ImagesPrune(s.Ctx, opts)
	if err != nil {
		return image.PruneReport{}, err
	}

	return pruneReport, nil
}

// UpdateContainerResources updates the resource limits of a running container.
//
// This function updates the memory, memory swap, and CPU shares limits of a container.
// It uses the Docker API's ContainerUpdate method which allows changing resource
// constraints on a running container without restarting it.
//
// Parameters:
//
//	containerID - the unique identifier of the container to update.
//	resources - the UpdateConfig containing the new resource limits.
//
// Returns:
//
//	container.ContainerUpdateOKBody - the response containing any warnings from the update.
//	error - an error if the update fails.
func (s *DockerService) UpdateContainerResources(containerID string, resources container.UpdateConfig) (container.ContainerUpdateOKBody, error) {
	return s.Cli.ContainerUpdate(s.Ctx, containerID, resources)
}
