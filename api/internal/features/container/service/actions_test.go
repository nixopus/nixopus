package service

import (
	"errors"
	"testing"

	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/stretchr/testify/require"
)

func TestStartContainer_successAndError(t *testing.T) {
	t.Parallel()
	log := logger.NewLogger()
	stub := &stubDockerRepository{}

	stub.startContainer = func(string, container.StartOptions) error { return nil }
	resp, err := StartContainer(stub, log, StartContainerOptions{ContainerID: "c1"})
	require.NoError(t, err)
	require.Equal(t, "started", resp.Data.Status)

	stub.startContainer = func(string, container.StartOptions) error { return errors.New("boom") }
	_, err = StartContainer(stub, log, StartContainerOptions{ContainerID: "c1"})
	require.Error(t, err)
}

func TestStopContainer_timeoutAndErrors(t *testing.T) {
	t.Parallel()
	log := logger.NewLogger()
	stub := &stubDockerRepository{}
	timeout := 5
	stub.stopContainer = func(id string, opts container.StopOptions) error {
		require.Equal(t, "c2", id)
		require.NotNil(t, opts.Timeout)
		require.Equal(t, 5, *opts.Timeout)
		return nil
	}
	resp, err := StopContainer(stub, log, StopContainerOptions{ContainerID: "c2", Timeout: &timeout})
	require.NoError(t, err)
	require.Equal(t, "stopped", resp.Data.Status)

	stub.stopContainer = func(string, container.StopOptions) error { return nil }
	resp, err = StopContainer(stub, log, StopContainerOptions{ContainerID: "c2"})
	require.NoError(t, err)
	require.Equal(t, "stopped", resp.Data.Status)

	stub.stopContainer = func(string, container.StopOptions) error { return errors.New("fail") }
	_, err = StopContainer(stub, log, StopContainerOptions{ContainerID: "c2"})
	require.Error(t, err)
}

func TestRestartContainer_timeoutAndErrors(t *testing.T) {
	t.Parallel()
	log := logger.NewLogger()
	stub := &stubDockerRepository{}
	timeout := 3
	stub.restartContainer = func(id string, opts container.StopOptions) error {
		require.Equal(t, "c3", id)
		require.NotNil(t, opts.Timeout)
		return nil
	}
	resp, err := RestartContainer(stub, log, RestartContainerOptions{ContainerID: "c3", Timeout: &timeout})
	require.NoError(t, err)
	require.Equal(t, "restarted", resp.Data.Status)

	stub.restartContainer = func(string, container.StopOptions) error { return errors.New("fail") }
	_, err = RestartContainer(stub, log, RestartContainerOptions{ContainerID: "c3"})
	require.Error(t, err)
}

func TestRemoveContainer_forceAndErrors(t *testing.T) {
	t.Parallel()
	log := logger.NewLogger()
	stub := &stubDockerRepository{}
	stub.removeContainer = func(id string, opts container.RemoveOptions) error {
		require.True(t, opts.Force)
		return nil
	}
	resp, err := RemoveContainer(stub, log, RemoveContainerOptions{ContainerID: "c4", Force: true})
	require.NoError(t, err)
	require.Equal(t, "removed", resp.Data.Status)

	stub.removeContainer = func(string, container.RemoveOptions) error { return errors.New("fail") }
	_, err = RemoveContainer(stub, log, RemoveContainerOptions{ContainerID: "c4"})
	require.Error(t, err)
}

func TestPruneBuildCache_successAndError(t *testing.T) {
	t.Parallel()
	log := logger.NewLogger()
	stub := &stubDockerRepository{}
	stub.pruneBuildCache = func(opts dockertypes.BuildCachePruneOptions) error {
		return nil
	}
	_, err := PruneBuildCache(stub, log, PruneBuildCacheOptions{All: true})
	require.NoError(t, err)

	var hit bool
	stub.pruneBuildCache = func(o dockertypes.BuildCachePruneOptions) error {
		hit = true
		require.True(t, o.All)
		return errors.New("prune failed")
	}
	_, err = PruneBuildCache(stub, log, PruneBuildCacheOptions{All: true})
	require.Error(t, err)
	require.True(t, hit)
}
