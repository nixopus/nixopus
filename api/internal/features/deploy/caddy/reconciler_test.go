package caddy

import (
	"testing"

	"github.com/docker/docker/api/types/swarm"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
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
