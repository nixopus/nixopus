package service_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"

	"crypto/sha256"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/telemetry/service"
	"github.com/nixopus/nixopus/api/internal/features/telemetry/storage"
	"github.com/nixopus/nixopus/api/internal/features/telemetry/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockTelemetryRepo is a test double for storage.TelemetryRepository.
type mockTelemetryRepo struct {
	createInstallFn func(event *types.CliInstallation) error
	capturedEvent   *types.CliInstallation
}

func (m *mockTelemetryRepo) CreateInstallEvent(event *types.CliInstallation) error {
	m.capturedEvent = event
	if m.createInstallFn != nil {
		return m.createInstallFn(event)
	}
	return nil
}

var _ storage.TelemetryRepository = (*mockTelemetryRepo)(nil)

// hashIPWithSalt reproduces the service's hashing logic for assertions.
func hashIPWithSalt(salt, ip string) string {
	h := sha256.New()
	h.Write([]byte(salt + ip))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func TestNewTelemetryService_defaultSalt(t *testing.T) {
	os.Unsetenv("TELEMETRY_IP_SALT")

	repo := &mockTelemetryRepo{}
	svc := service.NewTelemetryService(repo, context.Background(), logger.NewLogger())
	require.NotNil(t, svc)
}

func TestNewTelemetryService_customSalt(t *testing.T) {
	os.Setenv("TELEMETRY_IP_SALT", "my-custom-salt")
	defer os.Unsetenv("TELEMETRY_IP_SALT")

	repo := &mockTelemetryRepo{}
	svc := service.NewTelemetryService(repo, context.Background(), logger.NewLogger())
	require.NotNil(t, svc)
}

func TestTrackInstall_success(t *testing.T) {
	os.Setenv("TELEMETRY_IP_SALT", "test-salt")
	defer os.Unsetenv("TELEMETRY_IP_SALT")

	repo := &mockTelemetryRepo{}
	svc := service.NewTelemetryService(repo, context.Background(), logger.NewLogger())

	req := &types.TrackInstallRequest{
		EventType: "install_success",
		OS:        "ubuntu",
		Arch:      "amd64",
		Version:   "1.2.3",
		Duration:  60,
		Error:     "",
	}

	err := svc.TrackInstall(req, "203.0.113.42")
	require.NoError(t, err)

	require.NotNil(t, repo.capturedEvent)
	assert.Equal(t, req.EventType, repo.capturedEvent.EventType)
	assert.Equal(t, req.OS, repo.capturedEvent.OS)
	assert.Equal(t, req.Arch, repo.capturedEvent.Arch)
	assert.Equal(t, req.Version, repo.capturedEvent.Version)
	assert.Equal(t, req.Duration, repo.capturedEvent.Duration)
	assert.Equal(t, req.Error, repo.capturedEvent.Error)

	expectedHash := hashIPWithSalt("test-salt", "203.0.113.42")
	assert.Equal(t, expectedHash, repo.capturedEvent.IPHash)
	assert.NotEmpty(t, repo.capturedEvent.ID)
}

func TestTrackInstall_storageError(t *testing.T) {
	os.Unsetenv("TELEMETRY_IP_SALT")

	storageErr := errors.New("database unavailable")
	repo := &mockTelemetryRepo{
		createInstallFn: func(_ *types.CliInstallation) error {
			return storageErr
		},
	}
	svc := service.NewTelemetryService(repo, context.Background(), logger.NewLogger())

	req := &types.TrackInstallRequest{
		EventType: "install_failure",
		OS:        "debian",
		Arch:      "arm64",
		Version:   "0.1.0",
	}

	err := svc.TrackInstall(req, "10.0.0.1")
	require.Error(t, err)
	assert.Equal(t, storageErr, err)
}

func TestTrackInstall_defaultSaltHashesIP(t *testing.T) {
	os.Unsetenv("TELEMETRY_IP_SALT")

	repo := &mockTelemetryRepo{}
	svc := service.NewTelemetryService(repo, context.Background(), logger.NewLogger())

	req := &types.TrackInstallRequest{
		EventType: "install_started",
		OS:        "alpine",
		Arch:      "amd64",
		Version:   "2.0.0",
	}

	err := svc.TrackInstall(req, "192.168.1.1")
	require.NoError(t, err)

	const defaultSalt = "nixopus-telemetry-default-salt"
	expectedHash := hashIPWithSalt(defaultSalt, "192.168.1.1")
	assert.Equal(t, expectedHash, repo.capturedEvent.IPHash)
}
