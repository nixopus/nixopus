package service

import (
	"encoding/binary"
	"testing"

	"context"

	"github.com/google/uuid"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestTailLinesFromOrgLogTailSetting(t *testing.T) {
	t.Parallel()
	seven := 7
	require.Equal(t, 7, tailLinesFromOrgLogTailSetting(&seven))
	require.Equal(t, 100, tailLinesFromOrgLogTailSetting(nil))
}

func TestDecodeDockerLogs_variants(t *testing.T) {
	t.Parallel()

	t.Run("nil", func(t *testing.T) {
		require.Equal(t, "", decodeDockerLogs(nil))
	})

	t.Run("short_header", func(t *testing.T) {
		require.Equal(t, "", decodeDockerLogs([]byte{1, 0, 0, 0, 4}))
	})

	t.Run("stdout_and_stderr_concat", func(t *testing.T) {
		var frames []byte
		frames = appendFrame(frames, 1, []byte("out"))
		frames = appendFrame(frames, 2, []byte("err"))
		require.Equal(t, "outerr", decodeDockerLogs(frames))
	})

	t.Run("skip_unknown_stream", func(t *testing.T) {
		var frames []byte
		frames = appendFrame(frames, 7, []byte("x"))
		frames = appendFrame(frames, 1, []byte("y"))
		require.Equal(t, "y", decodeDockerLogs(frames))
	})

	t.Run("truncated_length", func(t *testing.T) {
		payload := byte(99)
		b := []byte{1, 0, 0, 0, 0, 0, 0, 50, payload}
		require.Equal(t, "", decodeDockerLogs(b))
	})

	t.Run("partial_header_at_end_loop_exit", func(t *testing.T) {
		payload := appendFrame(nil, 1, []byte("done"))
		partial := append(payload, byte(9))
		require.Equal(t, "done", decodeDockerLogs(partial))
	})
}

func appendFrame(b []byte, streamType byte, payload []byte) []byte {
	hdr := []byte{streamType, 0, 0, 0}
	lenb := make([]byte, 4)
	binary.BigEndian.PutUint32(lenb, uint32(len(payload)))
	hdr = append(hdr, lenb...)
	b = append(b, hdr...)
	return append(b, payload...)
}

func TestGetOrganizationSettings_nonDBPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("invalid_uuid", func(t *testing.T) {
		got := getOrganizationSettings(&shared_storage.Store{}, ctx, "nope-not-uuid")
		require.NotNil(t, got.ContainerLogTailLines)
		require.GreaterOrEqual(t, *got.ContainerLogTailLines, 1)
	})

	t.Run("nil_uuid_string", func(t *testing.T) {
		got := getOrganizationSettings(&shared_storage.Store{}, ctx, uuid.Nil.String())
		require.NotNil(t, got.ContainerLogTailLines)
	})

	t.Run("valid_uuid_nil_store_returns_defaults_without_db", func(t *testing.T) {
		got := getOrganizationSettings(nil, ctx, uuid.NewString())
		require.NotNil(t, got.ContainerLogTailLines)
	})

	t.Run("valid_uuid_nil_db_returns_defaults_without_db", func(t *testing.T) {
		st := &shared_storage.Store{}
		got := getOrganizationSettings(st, ctx, uuid.NewString())
		require.NotNil(t, got.ContainerLogTailLines)
	})
}
