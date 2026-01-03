package service

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/docker"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// ContainerLogsOptions contains options for fetching container logs
type ContainerLogsOptions struct {
	ContainerID    string
	OrganizationID string
	Follow         bool
	Tail           int
	Since          string
	Until          string
	Stdout         bool
	Stderr         bool
}

// GetContainerLogs fetches and decodes container logs.
// It handles organization settings, docker service initialization, and log decoding.
func GetContainerLogs(
	ctx context.Context,
	store *shared_storage.Store,
	dockerService *docker.DockerService,
	l logger.Logger,
	opts ContainerLogsOptions,
) (string, error) {
	// Get organization settings
	orgSettings := getOrganizationSettings(store, ctx, opts.OrganizationID)

	// Use default tail lines from settings if not provided
	tail := opts.Tail
	if tail == 0 {
		if orgSettings.ContainerLogTailLines != nil {
			tail = *orgSettings.ContainerLogTailLines
		} else {
			tail = 100 // Fallback default
		}
	}

	// Get container logs
	logsReader, err := dockerService.GetContainerLogs(opts.ContainerID, container.LogsOptions{
		Follow:     opts.Follow,
		Tail:       strconv.Itoa(tail),
		Since:      opts.Since,
		Until:      opts.Until,
		ShowStdout: opts.Stdout,
		ShowStderr: opts.Stderr,
	})
	if err != nil {
		l.Log(logger.Error, err.Error(), "")
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}

	// Read logs into buffer
	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logsReader)
	if err != nil {
		l.Log(logger.Error, err.Error(), "")
		return "", fmt.Errorf("failed to read container logs: %w", err)
	}

	// Decode Docker logs format
	decodedLogs := decodeDockerLogs(buf.Bytes())

	return decodedLogs, nil
}

// getOrganizationSettings retrieves organization settings with defaults
func getOrganizationSettings(store *shared_storage.Store, ctx context.Context, orgIDStr string) shared_types.OrganizationSettingsData {
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil || orgID == uuid.Nil {
		return shared_types.DefaultOrganizationSettingsData()
	}

	orgStore := storage.OrganizationStore{DB: store.DB, Ctx: ctx}
	settings, err := orgStore.GetOrganizationSettings(orgID.String())
	if err != nil || settings == nil {
		return shared_types.DefaultOrganizationSettingsData()
	}

	// Merge with defaults to ensure all fields are set
	defaults := shared_types.DefaultOrganizationSettingsData()
	result := shared_types.OrganizationSettingsData{
		WebsocketReconnectAttempts:       settings.Settings.WebsocketReconnectAttempts,
		WebsocketReconnectInterval:       settings.Settings.WebsocketReconnectInterval,
		ApiRetryAttempts:                 settings.Settings.ApiRetryAttempts,
		DisableApiCache:                  settings.Settings.DisableApiCache,
		ContainerLogTailLines:            defaults.ContainerLogTailLines,
		ContainerDefaultRestartPolicy:    defaults.ContainerDefaultRestartPolicy,
		ContainerStopTimeout:             defaults.ContainerStopTimeout,
		ContainerAutoPruneDanglingImages: defaults.ContainerAutoPruneDanglingImages,
		ContainerAutoPruneBuildCache:     defaults.ContainerAutoPruneBuildCache,
	}

	if settings.Settings.ContainerLogTailLines != nil {
		result.ContainerLogTailLines = settings.Settings.ContainerLogTailLines
	}
	if settings.Settings.ContainerDefaultRestartPolicy != nil {
		result.ContainerDefaultRestartPolicy = settings.Settings.ContainerDefaultRestartPolicy
	}
	if settings.Settings.ContainerStopTimeout != nil {
		result.ContainerStopTimeout = settings.Settings.ContainerStopTimeout
	}
	if settings.Settings.ContainerAutoPruneDanglingImages != nil {
		result.ContainerAutoPruneDanglingImages = settings.Settings.ContainerAutoPruneDanglingImages
	}
	if settings.Settings.ContainerAutoPruneBuildCache != nil {
		result.ContainerAutoPruneBuildCache = settings.Settings.ContainerAutoPruneBuildCache
	}

	return result
}

// decodeDockerLogs decodes Docker's log format (8-byte header + payload)
func decodeDockerLogs(data []byte) string {
	var result bytes.Buffer
	offset := 0

	for offset < len(data) {
		if offset+8 > len(data) {
			break
		}

		streamType := data[offset]
		length := binary.BigEndian.Uint32(data[offset+4 : offset+8])
		offset += 8

		if offset+int(length) > len(data) {
			break
		}

		if streamType == 1 || streamType == 2 {
			result.Write(data[offset : offset+int(length)])
		}
		offset += int(length)
	}

	return result.String()
}
