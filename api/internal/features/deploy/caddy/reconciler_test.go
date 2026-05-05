package caddy

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/api/types/volume"
	client "github.com/docker/docker/client"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/deploy/docker"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_upstreamsEqual(t *testing.T) {
	assert.True(t, upstreamsEqual(nil, nil))
	a := []string{"x:1", "y:2"}
	b := []string{"x:1", "y:2"}
	assert.True(t, upstreamsEqual(a, b))
	copyB := append([]string(nil), b...)
	assert.True(t, upstreamsEqual(a, copyB))

	assert.False(t, upstreamsEqual([]string{"x:1", "y:2"}, []string{"x:1", "y:2", "z:3"}))
	assert.False(t, upstreamsEqual([]string{"a", "b"}, []string{"b", "a"}))
}

func Test_extractPublishedPort(t *testing.T) {
	name := "websvc"
	pub := uint32(7443)

	svc := swarm.Service{}
	svc.Spec.Annotations.Name = name
	svc.Endpoint.Ports = []swarm.PortConfig{{PublishedPort: pub}}

	got, err := extractPublishedPort(svc)
	require.NoError(t, err)
	assert.Equal(t, int(pub), got)

	specOnly := swarm.Service{}
	specOnly.Spec.Annotations.Name = name
	specOnly.Spec.EndpointSpec = &swarm.EndpointSpec{
		Ports: []swarm.PortConfig{{PublishedPort: 9000}},
	}
	got2, err := extractPublishedPort(specOnly)
	require.NoError(t, err)
	assert.Equal(t, 9000, got2)

	noPort := swarm.Service{}
	noPort.Spec.Annotations.Name = name
	_, err = extractPublishedPort(noPort)
	require.Error(t, err)
}

func TestEnqueueReconcile_queueNotInitialized(t *testing.T) {
	origQ, origT := ReconcileQueue, TaskCaddyReconcile
	ReconcileQueue, TaskCaddyReconcile = nil, nil
	t.Cleanup(func() {
		ReconcileQueue, TaskCaddyReconcile = origQ, origT
	})

	assert.Error(t, EnqueueReconcile(uuid.New()))
}

func TestReconciler_orgLock_returnsStableMutexPerOrg(t *testing.T) {
	l := logger.NewLogger()
	stubRepo := &deployRepoTestStub{}
	r := NewReconciler(stubRepo, l)
	a := uuid.New()
	b := uuid.New()

	muA := r.orgLock(a)
	muB := r.orgLock(b)
	assert.Same(t, muA, r.orgLock(a))
	assert.NotSame(t, muA, muB)
}

func TestReconciler_fullRebuild_emptyDesired(t *testing.T) {
	l := logger.NewLogger()
	stubRepo := &deployRepoTestStub{}
	r := NewReconciler(stubRepo, l)

	ctx := t.Context()
	got, err := r.fullRebuild(ctx, nil)
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Empty(t, got.Added)

	got2, err := r.fullRebuild(ctx, []DomainRoute{})
	require.NoError(t, err)
	assert.Empty(t, got2.Added)
}

// ---- buildDeployDomains ---------------------------------------------------

func TestReconciler_buildDeployDomains_storageError(t *testing.T) {
	r := NewReconciler(&deployRepoTestStub{deployedAppsErr: errors.New("db down")}, logger.NewLogger())
	_, err := r.buildDeployDomains(context.Background(), uuid.New(), "proxy.internal")
	require.Error(t, err)
}

func TestReconciler_buildDeployDomains_appNoDomains(t *testing.T) {
	app := shared_types.Application{ID: uuid.New(), BuildPack: shared_types.DockerFile}
	r := NewReconciler(&deployRepoTestStub{deployedApps: []shared_types.Application{app}}, logger.NewLogger())
	routes, err := r.buildDeployDomains(context.Background(), uuid.New(), "proxy.internal")
	require.NoError(t, err)
	assert.Empty(t, routes)
}

func TestReconciler_buildDeployDomains_dockerCompose(t *testing.T) {
	port := 3000
	d := shared_types.ApplicationDomain{Domain: "compose.example", Port: &port}
	app := shared_types.Application{
		ID:        uuid.New(),
		BuildPack: shared_types.DockerCompose,
		Domains:   []*shared_types.ApplicationDomain{&d},
	}
	r := NewReconciler(&deployRepoTestStub{deployedApps: []shared_types.Application{app}}, logger.NewLogger())
	routes, err := r.buildDeployDomains(context.Background(), uuid.New(), "proxy.internal")
	require.NoError(t, err)
	require.Len(t, routes, 1)
	assert.Equal(t, "compose.example", routes[0].Domain)
	assert.Equal(t, "proxy.internal:3000", routes[0].UpstreamDial)
}

func TestReconciler_buildDeployDomains_dockerCompose_noPort(t *testing.T) {
	// Domain with no ComposeService and no Port → ResolvePort() == 0 → skipped.
	d := shared_types.ApplicationDomain{Domain: "orphan.compose"}
	app := shared_types.Application{
		ID:        uuid.New(),
		BuildPack: shared_types.DockerCompose,
		Domains:   []*shared_types.ApplicationDomain{&d},
	}
	r := NewReconciler(&deployRepoTestStub{deployedApps: []shared_types.Application{app}}, logger.NewLogger())
	routes, err := r.buildDeployDomains(context.Background(), uuid.New(), "proxy.internal")
	require.NoError(t, err)
	assert.Empty(t, routes)
}

// ---- getPublishedPort via hook --------------------------------------------

func TestReconciler_getPublishedPort_hookError(t *testing.T) {
	orig := reconcileGetPublishedPortHook
	reconcileGetPublishedPortHook = func(_ *Reconciler, _ context.Context, _ string) (int, error) {
		return 0, errors.New("hook port error")
	}
	t.Cleanup(func() { reconcileGetPublishedPortHook = orig })

	r := NewReconciler(&deployRepoTestStub{}, logger.NewLogger())
	_, err := r.getPublishedPort(context.Background(), "my-svc")
	require.Error(t, err)
}

func TestReconciler_getPublishedPort_dockerServiceContextError(t *testing.T) {
	orig := reconcileGetPublishedPortHook
	reconcileGetPublishedPortHook = nil
	t.Cleanup(func() { reconcileGetPublishedPortHook = orig })

	origDocker := getDockerServiceFromContext
	getDockerServiceFromContext = func(_ context.Context) (docker.DockerRepository, error) {
		return nil, errors.New("no docker context")
	}
	t.Cleanup(func() { getDockerServiceFromContext = origDocker })

	r := NewReconciler(&deployRepoTestStub{}, logger.NewLogger())
	_, err := r.getPublishedPort(context.Background(), "svc")
	require.Error(t, err)
}

// ---- getSSHHostForOrg ----------------------------------------------------

func TestGetSSHHostForOrg_upstreamHostError(t *testing.T) {
	orig := getSSHManagerFromContextForCaddy
	getSSHManagerFromContextForCaddy = func(_ context.Context) (*ssh.SSHManager, error) {
		// Return an empty manager; GetUpstreamHost will fail (no default client).
		return ssh.NewSSHManager(), nil
	}
	t.Cleanup(func() { getSSHManagerFromContextForCaddy = orig })

	origHook := getSSHHostForOrgHook
	getSSHHostForOrgHook = nil
	t.Cleanup(func() { getSSHHostForOrgHook = origHook })

	_, err := getSSHHostForOrg(context.Background())
	require.Error(t, err)
}

// ---- buildDesiredState: extension-domains Redis error --------------------

func TestReconciler_buildDesiredState_redisError(t *testing.T) {
	// Without Redis initialized, GetExtensionDomains returns an error.
	// buildDesiredState should log a warning and continue (not fail).
	r := NewReconciler(&deployRepoTestStub{}, logger.NewLogger())
	// Make getSSHHostForOrgHook available so orgLock and the function can be called directly.
	// We call buildDesiredState directly – no hook needed.
	routes, err := r.buildDesiredState(context.Background(), uuid.New(), "proxy.internal")
	require.NoError(t, err)
	assert.Empty(t, routes)
}

// ---- NewReconcilerDaemon and handleReconcileTask --------------------------

func TestNewReconcilerDaemon(t *testing.T) {
	d := NewReconcilerDaemon(&deployRepoTestStub{}, logger.NewLogger(), time.Second,
		func(context.Context) ([]uuid.UUID, error) { return nil, nil })
	require.NotNil(t, d)
	require.NotNil(t, d.Reconciler())
}

func TestHandleReconcileTask_invalidOrgID(t *testing.T) {
	d := &ReconcilerDaemon{
		reconciler: NewReconciler(&deployRepoTestStub{}, logger.NewLogger()),
		logger:     logger.NewLogger(),
	}
	err := d.handleReconcileTask(context.Background(), CaddyTaskPayload{OrganizationID: "not-a-uuid"})
	require.Error(t, err)
}

func TestHandleReconcileTask_validOrgID_reconcileError(t *testing.T) {
	origHook := getSSHHostForOrgHook
	getSSHHostForOrgHook = func(context.Context) (string, error) {
		return "", errors.New("no ssh")
	}
	t.Cleanup(func() { getSSHHostForOrgHook = origHook })

	d := &ReconcilerDaemon{
		reconciler: NewReconciler(&deployRepoTestStub{}, logger.NewLogger()),
		logger:     logger.NewLogger(),
	}
	err := d.handleReconcileTask(context.Background(), CaddyTaskPayload{OrganizationID: uuid.New().String()})
	require.Error(t, err)
}

// ---- dockerRepoStub satisfies docker.DockerRepository for getPublishedPort tests ----

type dockerRepoStub struct {
	svcName string
	svc     *swarm.Service
	svcErr  error
}

func (s *dockerRepoStub) GetServiceByName(name string) (*swarm.Service, error) {
	return s.svc, s.svcErr
}

// All remaining methods are no-ops.
func (s *dockerRepoStub) ListAllContainers() ([]container.Summary, error) { return nil, nil }
func (s *dockerRepoStub) ListContainers(container.ListOptions) ([]container.Summary, error) {
	return nil, nil
}
func (s *dockerRepoStub) ListAllImages(image.ListOptions) []image.Summary       { return nil }
func (s *dockerRepoStub) StopContainer(string, container.StopOptions) error     { return nil }
func (s *dockerRepoStub) RemoveContainer(string, container.RemoveOptions) error { return nil }
func (s *dockerRepoStub) StartContainer(string, container.StartOptions) error   { return nil }
func (s *dockerRepoStub) GetContainerLogs(string, container.LogsOptions) (io.Reader, error) {
	return nil, nil
}
func (s *dockerRepoStub) GetContainerById(string) (container.InspectResponse, error) {
	return container.InspectResponse{}, nil
}
func (s *dockerRepoStub) GetImageById(string, client.ImageInspectOption) (image.InspectResponse, error) {
	return image.InspectResponse{}, nil
}
func (s *dockerRepoStub) ImagePull(context.Context, string, image.PullOptions) (io.ReadCloser, error) {
	return nil, nil
}
func (s *dockerRepoStub) BuildImage(types.ImageBuildOptions, io.Reader) (types.ImageBuildResponse, error) {
	return types.ImageBuildResponse{}, nil
}
func (s *dockerRepoStub) CreateContainer(container.Config, container.HostConfig, network.NetworkingConfig, string) (container.CreateResponse, error) {
	return container.CreateResponse{}, nil
}
func (s *dockerRepoStub) ContainerLogs(context.Context, string, container.LogsOptions) (io.ReadCloser, error) {
	return nil, nil
}
func (s *dockerRepoStub) RestartContainer(string, container.StopOptions) error { return nil }
func (s *dockerRepoStub) UpdateContainerResources(string, container.UpdateConfig) (container.ContainerUpdateOKBody, error) {
	return container.ContainerUpdateOKBody{}, nil
}
func (s *dockerRepoStub) ComposeUp(string, map[string]string, ...string) (string, error) {
	return "", nil
}
func (s *dockerRepoStub) ComposeUpWithCallback(string, map[string]string, func(string), ...string) (string, error) {
	return "", nil
}
func (s *dockerRepoStub) ComposeDown(string, map[string]string, ...string) error { return nil }
func (s *dockerRepoStub) ComposeDownWithCallback(string, map[string]string, func(string), ...string) error {
	return nil
}
func (s *dockerRepoStub) ComposeRestart(string, map[string]string, func(string), ...string) error {
	return nil
}
func (s *dockerRepoStub) ComposeBuild(string, map[string]string, ...string) error { return nil }
func (s *dockerRepoStub) RemoveImage(string, image.RemoveOptions) error           { return nil }
func (s *dockerRepoStub) PruneBuildCache(types.BuildCachePruneOptions) error      { return nil }
func (s *dockerRepoStub) PruneImages(filters.Args) (image.PruneReport, error) {
	return image.PruneReport{}, nil
}
func (s *dockerRepoStub) InitCluster() error                                          { return nil }
func (s *dockerRepoStub) JoinCluster() error                                          { return nil }
func (s *dockerRepoStub) LeaveCluster(bool) error                                     { return nil }
func (s *dockerRepoStub) GetClusterInfo() (swarm.ClusterInfo, error)                  { return swarm.ClusterInfo{}, nil }
func (s *dockerRepoStub) GetClusterNodes() ([]swarm.Node, error)                      { return nil, nil }
func (s *dockerRepoStub) GetClusterServices() ([]swarm.Service, error)                { return nil, nil }
func (s *dockerRepoStub) GetClusterTasks() ([]swarm.Task, error)                      { return nil, nil }
func (s *dockerRepoStub) GetTasksByServiceID(string) ([]swarm.Task, error)            { return nil, nil }
func (s *dockerRepoStub) GetClusterSecrets() ([]swarm.Secret, error)                  { return nil, nil }
func (s *dockerRepoStub) GetClusterConfigs() ([]swarm.Config, error)                  { return nil, nil }
func (s *dockerRepoStub) GetClusterVolumes() ([]*volume.Volume, error)                { return nil, nil }
func (s *dockerRepoStub) GetClusterNetworks() ([]network.Summary, error)              { return nil, nil }
func (s *dockerRepoStub) UpdateNodeAvailability(string, swarm.NodeAvailability) error { return nil }
func (s *dockerRepoStub) ScaleService(string, uint64, string) error                   { return nil }
func (s *dockerRepoStub) ListenEvents(events.ListOptions) (<-chan events.Message, <-chan error) {
	return nil, nil
}
func (s *dockerRepoStub) GetServiceHealth(swarm.Service) (int, int, error)      { return 0, 0, nil }
func (s *dockerRepoStub) GetTaskHealth(swarm.Task) swarm.TaskState              { return "" }
func (s *dockerRepoStub) CreateService(swarm.Service) error                     { return nil }
func (s *dockerRepoStub) UpdateService(string, swarm.ServiceSpec, string) error { return nil }
func (s *dockerRepoStub) DeleteService(string) error                            { return nil }
func (s *dockerRepoStub) RollbackService(string) error                          { return nil }
func (s *dockerRepoStub) GetServiceByID(string) (swarm.Service, error)          { return swarm.Service{}, nil }
func (s *dockerRepoStub) GetServiceByLabel(string, string) (*swarm.Service, error) {
	return nil, nil
}

var _ docker.DockerRepository = (*dockerRepoStub)(nil)

// ---- getPublishedPort via docker stub ------------------------------------

func TestReconciler_getPublishedPort_getServiceByNameError(t *testing.T) {
	origHook := reconcileGetPublishedPortHook
	reconcileGetPublishedPortHook = nil
	t.Cleanup(func() { reconcileGetPublishedPortHook = origHook })

	stub := &dockerRepoStub{svcErr: errors.New("swarm unreachable")}
	origDocker := getDockerServiceFromContext
	getDockerServiceFromContext = func(_ context.Context) (docker.DockerRepository, error) {
		return stub, nil
	}
	t.Cleanup(func() { getDockerServiceFromContext = origDocker })

	r := NewReconciler(&deployRepoTestStub{}, logger.NewLogger())
	_, err := r.getPublishedPort(context.Background(), "svc")
	require.Error(t, err)
}

func TestReconciler_getPublishedPort_nilService(t *testing.T) {
	origHook := reconcileGetPublishedPortHook
	reconcileGetPublishedPortHook = nil
	t.Cleanup(func() { reconcileGetPublishedPortHook = origHook })

	stub := &dockerRepoStub{svc: nil, svcErr: nil}
	origDocker := getDockerServiceFromContext
	getDockerServiceFromContext = func(_ context.Context) (docker.DockerRepository, error) {
		return stub, nil
	}
	t.Cleanup(func() { getDockerServiceFromContext = origDocker })

	r := NewReconciler(&deployRepoTestStub{}, logger.NewLogger())
	_, err := r.getPublishedPort(context.Background(), "svc")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found in swarm")
}

func TestReconciler_getPublishedPort_withPublishedPort(t *testing.T) {
	origHook := reconcileGetPublishedPortHook
	reconcileGetPublishedPortHook = nil
	t.Cleanup(func() { reconcileGetPublishedPortHook = origHook })

	svc := &swarm.Service{}
	svc.Spec.Annotations.Name = "my-svc"
	svc.Endpoint.Ports = []swarm.PortConfig{{PublishedPort: 9000}}

	stub := &dockerRepoStub{svc: svc}
	origDocker := getDockerServiceFromContext
	getDockerServiceFromContext = func(_ context.Context) (docker.DockerRepository, error) {
		return stub, nil
	}
	t.Cleanup(func() { getDockerServiceFromContext = origDocker })

	r := NewReconciler(&deployRepoTestStub{}, logger.NewLogger())
	port, err := r.getPublishedPort(context.Background(), "my-svc")
	require.NoError(t, err)
	assert.Equal(t, 9000, port)
}
