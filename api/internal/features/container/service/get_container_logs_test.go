package service

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	"github.com/stretchr/testify/require"
)

type errLogReader struct{}

func (errLogReader) Read([]byte) (int, error) { return 0, errors.New("read failed") }

func TestGetContainerLogs_docker_error(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	log := logger.NewLogger()
	st := &shared_storage.Store{}
	stub := &stubDockerRepository{}
	stub.getContainerLogs = func(string, container.LogsOptions) (io.Reader, error) {
		return nil, errors.New("daemon down")
	}

	_, err := GetContainerLogs(ctx, st, stub, log, ContainerLogsOptions{
		ContainerID:    "c1",
		OrganizationID: "not-a-uuid",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get container logs")
}

func TestGetContainerLogs_read_error(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	log := logger.NewLogger()
	st := &shared_storage.Store{}
	stub := &stubDockerRepository{}
	stub.getContainerLogs = func(cid string, opts container.LogsOptions) (io.Reader, error) {
		require.Equal(t, "c-read", cid)
		require.Equal(t, "5", opts.Tail)
		return errLogReader{}, nil
	}

	_, err := GetContainerLogs(ctx, st, stub, log, ContainerLogsOptions{
		ContainerID:    "c-read",
		OrganizationID: "also-bad",
		Tail:           5,
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read container logs")
}

func TestGetContainerLogs_success_decodes_stream(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	log := logger.NewLogger()
	st := &shared_storage.Store{}
	payload := appendFrame(nil, 1, []byte("HELLO"))

	stub := &stubDockerRepository{}
	stub.getContainerLogs = func(string, container.LogsOptions) (io.Reader, error) {
		return strings.NewReader(string(payload)), nil
	}

	out, err := GetContainerLogs(ctx, st, stub, log, ContainerLogsOptions{
		ContainerID:    "cok",
		OrganizationID: "bad-org-no-db",
		Stdout:         true,
		Stderr:         true,
	})
	require.NoError(t, err)
	require.Equal(t, "HELLO", out)
}

func TestGetContainerLogs_explicit_tail_skips_org_settings_branch(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	log := logger.NewLogger()
	st := &shared_storage.Store{}
	var sawTail string
	stub := &stubDockerRepository{}
	stub.getContainerLogs = func(_ string, opts container.LogsOptions) (io.Reader, error) {
		sawTail = opts.Tail
		return strings.NewReader(""), nil
	}

	_, err := GetContainerLogs(ctx, st, stub, log, ContainerLogsOptions{
		ContainerID:    "c",
		OrganizationID: "",
		Tail:           42,
	})
	require.NoError(t, err)
	require.Equal(t, "42", sawTail)
}

func TestGetContainerLogs_tail_zero_uses_default_setting(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	log := logger.NewLogger()
	st := &shared_storage.Store{}
	stub := &stubDockerRepository{}
	stub.getContainerLogs = func(_ string, opts container.LogsOptions) (io.Reader, error) {
		require.Equal(t, "100", opts.Tail)
		return strings.NewReader(""), nil
	}

	_, err := GetContainerLogs(ctx, st, stub, log, ContainerLogsOptions{
		ContainerID:    "c",
		OrganizationID: "not-valid-uuid",
		Tail:           0,
	})
	require.NoError(t, err)
}
