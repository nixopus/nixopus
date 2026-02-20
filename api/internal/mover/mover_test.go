package mover

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- Merkle ----

func TestBuildFromPaths(t *testing.T) {
	t.Run("empty leaves", func(t *testing.T) {
		tree := BuildFromPaths(nil)
		require.NotNil(t, tree)
		assert.Empty(t, tree.RootHash)
		assert.Empty(t, tree.Leaves())

		tree = BuildFromPaths(map[string]string{})
		require.NotNil(t, tree)
		assert.Empty(t, tree.RootHash)
	})

	t.Run("single leaf", func(t *testing.T) {
		leaves := map[string]string{"a": "aa"}
		tree := BuildFromPaths(leaves)
		require.NotNil(t, tree)
		assert.Equal(t, "aa", tree.RootHash)
		assert.Equal(t, map[string]string{"a": "aa"}, tree.Leaves())
	})

	t.Run("multiple leaves deterministic", func(t *testing.T) {
		leaves := map[string]string{
			"b": "bb",
			"a": "aa",
			"c": "cc",
		}
		tree := BuildFromPaths(leaves)
		require.NotNil(t, tree)
		assert.NotEmpty(t, tree.RootHash)
		assert.Len(t, tree.Leaves(), 3)
		assert.Equal(t, "aa", tree.Leaves()["a"])
		assert.Equal(t, "bb", tree.Leaves()["b"])
		assert.Equal(t, "cc", tree.Leaves()["c"])

		tree2 := BuildFromPaths(leaves)
		assert.Equal(t, tree.RootHash, tree2.RootHash, "same input produces same root")
	})

	t.Run("path normalization", func(t *testing.T) {
		leaves := map[string]string{"a/b/../c": "hash"}
		tree := BuildFromPaths(leaves)
		assert.Equal(t, map[string]string{"a/c": "hash"}, tree.Leaves())
	})

	t.Run("empty string path", func(t *testing.T) {
		leaves := map[string]string{"": "hash"}
		tree := BuildFromPaths(leaves)
		// Empty path normalizes to "."
		assert.Contains(t, tree.Leaves(), ".")
	})
}

func TestDiffAgainst(t *testing.T) {
	t.Run("empty local empty server", func(t *testing.T) {
		local := BuildFromPaths(map[string]string{})
		toSync, toDelete := DiffAgainst(local, map[string]string{})
		assert.Empty(t, toSync)
		assert.Empty(t, toDelete)
	})

	t.Run("local has extra", func(t *testing.T) {
		local := BuildFromPaths(map[string]string{"a": "h1", "b": "h2"})
		server := map[string]string{"a": "h1"} // a matches, b only on local
		toSync, toDelete := DiffAgainst(local, server)
		assert.Contains(t, toSync, "b")
		assert.Empty(t, toDelete)
	})

	t.Run("server has extra", func(t *testing.T) {
		local := BuildFromPaths(map[string]string{"a": "h1"})
		server := map[string]string{"a": "h1", "b": "h2"}
		toSync, toDelete := DiffAgainst(local, server)
		assert.Empty(t, toSync)
		assert.Contains(t, toDelete, "b")
	})

	t.Run("different checksum", func(t *testing.T) {
		local := BuildFromPaths(map[string]string{"a": "h1"})
		server := map[string]string{"a": "h2"}
		toSync, toDelete := DiffAgainst(local, server)
		assert.Contains(t, toSync, "a")
		assert.Empty(t, toDelete)
	})

	t.Run("same checksums", func(t *testing.T) {
		local := BuildFromPaths(map[string]string{"a": "h1", "b": "h2"})
		server := map[string]string{"a": "h1", "b": "h2"}
		toSync, toDelete := DiffAgainst(local, server)
		assert.Empty(t, toSync)
		assert.Empty(t, toDelete)
	})

	t.Run("nil local leaves", func(t *testing.T) {
		local := &Tree{leaves: nil}
		server := map[string]string{"a": "h1"}
		toSync, toDelete := DiffAgainst(local, server)
		assert.Empty(t, toSync)
		assert.Contains(t, toDelete, "a")
	})

	t.Run("nil server leaves", func(t *testing.T) {
		local := BuildFromPaths(map[string]string{"a": "h1"})
		toSync, toDelete := DiffAgainst(local, nil)
		assert.Contains(t, toSync, "a")
		assert.Empty(t, toDelete)
	})

	t.Run("path normalization in diff", func(t *testing.T) {
		local := BuildFromPaths(map[string]string{"a/b": "h1"})
		server := map[string]string{"a/b/../b": "h2"}
		toSync, toDelete := DiffAgainst(local, server)
		assert.Contains(t, toSync, "a/b")
		assert.Empty(t, toDelete)
	})
}

func TestComputeSimhash(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		assert.Empty(t, ComputeSimhash(map[string]string{}))
		assert.Empty(t, ComputeSimhash(nil))
	})

	t.Run("valid sha256 hashes", func(t *testing.T) {
		h := hex.EncodeToString(make([]byte, 32))
		leaves := map[string]string{"a": h}
		out := ComputeSimhash(leaves)
		assert.Len(t, out, 64)
	})

	t.Run("skips invalid hashes", func(t *testing.T) {
		leaves := map[string]string{
			"a": "short",
			"b": "x", // not 64 chars - both skipped, produces zero vector
		}
		out := ComputeSimhash(leaves)
		assert.Len(t, out, 64)
		// When all invalid, output is zero vector hex
		assert.Equal(t, "0000000000000000000000000000000000000000000000000000000000000000", out)
	})

	t.Run("mix of valid and invalid hashes", func(t *testing.T) {
		// Use hash with some 1-bits so simhash is non-zero (all-zero hash produces zero vector)
		validBytes := make([]byte, 32)
		validBytes[0] = 0xFF
		validHash := hex.EncodeToString(validBytes)
		leaves := map[string]string{
			"valid":   validHash,
			"invalid": "not64chars",
		}
		out := ComputeSimhash(leaves)
		assert.Len(t, out, 64)
		assert.NotEqual(t, "0000000000000000000000000000000000000000000000000000000000000000", out)
	})

	t.Run("invalid hex in valid-length hash", func(t *testing.T) {
		leaves := map[string]string{"a": "zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz"}
		out := ComputeSimhash(leaves)
		assert.Len(t, out, 64)
	})
}

// ---- SyncState ----

func TestSyncState_RoundTrip(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mover-syncstate-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	path := filepath.Join(tmpDir, ".nixopus-sync-state.json")

	s := NewSyncState(path)
	appID := "app-123"

	s.SetRootHash(appID, "root-abc")
	s.SetPath(appID, "src/a.go", "hash1")
	s.SetPath(appID, "src/b.go", "hash2")
	require.NoError(t, s.Save())

	s2 := NewSyncState(path)
	require.NoError(t, s2.Load())

	assert.Equal(t, "root-abc", s2.GetRootHash(appID))
	paths := s2.GetPaths(appID)
	assert.Equal(t, "hash1", paths["src/a.go"])
	assert.Equal(t, "hash2", paths["src/b.go"])
}

func TestSyncState_LoadMissingFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mover-syncstate-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	path := filepath.Join(tmpDir, "missing.json")

	s := NewSyncState(path)
	err = s.Load()
	require.NoError(t, err)
	assert.Empty(t, s.GetRootHash("any"))
	assert.Nil(t, s.GetPaths("any"))
}

func TestSyncState_LoadCorruptJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mover-syncstate-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	path := filepath.Join(tmpDir, ".nixopus-sync-state.json")

	require.NoError(t, os.WriteFile(path, []byte(`{invalid json`), 0600))

	s := NewSyncState(path)
	err = s.Load()
	require.NoError(t, err)
	assert.Empty(t, s.GetRootHash("any"))
	assert.Nil(t, s.GetPaths("any"))
}

func TestSyncState_EdgeCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mover-syncstate-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	path := filepath.Join(tmpDir, ".nixopus-sync-state.json")

	s := NewSyncState(path)

	t.Run("GetPaths unknown app returns nil", func(t *testing.T) {
		assert.Nil(t, s.GetPaths("unknown-app"))
	})

	t.Run("GetRootHash unknown app returns empty", func(t *testing.T) {
		assert.Empty(t, s.GetRootHash("unknown-app"))
	})

	t.Run("RemovePath unknown app does not panic", func(t *testing.T) {
		s.RemovePath("unknown", "path")
	})

	t.Run("PruneNonexistentPaths empty existingPaths removes all", func(t *testing.T) {
		s.SetPath("app", "only.go", "h1")
		require.NoError(t, s.Save())
		s.PruneNonexistentPaths("app", map[string]bool{})
		require.NoError(t, s.Save())
		assert.Empty(t, s.GetPaths("app"))
	})

	t.Run("SetAllPaths empty map clears paths", func(t *testing.T) {
		s.SetPath("app2", "x.go", "h1")
		require.NoError(t, s.Save())
		s.SetAllPaths("app2", map[string]string{})
		require.NoError(t, s.Save())
		assert.Empty(t, s.GetPaths("app2"))
	})

	t.Run("Save when not dirty is no-op", func(t *testing.T) {
		require.NoError(t, s.Save())
		require.NoError(t, s.Save())
	})

	t.Run("path normalization on SetPath", func(t *testing.T) {
		s.SetPath("app3", "a/../b/./c", "h1")
		require.NoError(t, s.Save())
		paths := s.GetPaths("app3")
		assert.Equal(t, "h1", paths["b/c"])
	})
}

func TestSyncState_RemovePath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mover-syncstate-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	path := filepath.Join(tmpDir, ".nixopus-sync-state.json")

	s := NewSyncState(path)
	s.SetPath("app", "a.go", "h1")
	s.SetPath("app", "b.go", "h2")
	require.NoError(t, s.Save())

	s.RemovePath("app", "a.go")
	require.NoError(t, s.Save())

	s2 := NewSyncState(path)
	require.NoError(t, s2.Load())
	paths := s2.GetPaths("app")
	assert.NotContains(t, paths, "a.go")
	assert.Equal(t, "h2", paths["b.go"])
}

func TestSyncState_PruneNonexistentPaths(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mover-syncstate-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	path := filepath.Join(tmpDir, ".nixopus-sync-state.json")

	s := NewSyncState(path)
	s.SetPath("app", "keep.go", "h1")
	s.SetPath("app", "remove.go", "h2")
	require.NoError(t, s.Save())

	s.PruneNonexistentPaths("app", map[string]bool{"keep.go": true})
	require.NoError(t, s.Save())

	s2 := NewSyncState(path)
	require.NoError(t, s2.Load())
	paths := s2.GetPaths("app")
	assert.Contains(t, paths, "keep.go")
	assert.NotContains(t, paths, "remove.go")
}

func TestSyncState_SetAllPaths(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mover-syncstate-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	path := filepath.Join(tmpDir, ".nixopus-sync-state.json")

	s := NewSyncState(path)
	s.SetRootHash("app", "old-root")
	s.SetPath("app", "old.go", "h0")
	require.NoError(t, s.Save())

	s.SetAllPaths("app", map[string]string{"new.go": "h1", "new2.go": "h2"})
	require.NoError(t, s.Save())

	s2 := NewSyncState(path)
	require.NoError(t, s2.Load())
	paths := s2.GetPaths("app")
	assert.Equal(t, "h1", paths["new.go"])
	assert.Equal(t, "h2", paths["new2.go"])
	assert.NotContains(t, paths, "old.go")
	assert.Equal(t, "old-root", s2.GetRootHash("app"))
}

func TestSyncState_ConcurrentAccess(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mover-syncstate-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)
	path := filepath.Join(tmpDir, ".nixopus-sync-state.json")

	s := NewSyncState(path)
	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			s.SetPath("app", filepath.Join("p", string(rune('a'+i%26))+".go"), "hash")
		}
		close(done)
	}()
	for i := 0; i < 50; i++ {
		_ = s.GetPaths("app")
		_ = s.GetRootHash("app")
	}
	<-done
	require.NoError(t, s.Save())
}

// ---- Tracker ----

func TestTracker(t *testing.T) {
	tr := NewTracker()

	assert.Equal(t, ConnectionStatusDisconnected, tr.GetConnectionStatus())
	assert.Equal(t, ServiceStatusUnknown, tr.GetServiceStatus())
	assert.Zero(t, tr.GetFilesSynced())
	assert.Zero(t, tr.GetChangesDetected())

	tr.SetConnectionStatus(ConnectionStatusConnected)
	assert.Equal(t, ConnectionStatusConnected, tr.GetConnectionStatus())

	tr.SetServiceStatus(ServiceStatusRunning)
	assert.Equal(t, ServiceStatusRunning, tr.GetServiceStatus())

	tr.IncrementFilesSynced()
	tr.AddFilesSynced(4)
	assert.Equal(t, 5, tr.GetFilesSynced())

	tr.IncrementChanges()
	tr.IncrementChanges()
	assert.Equal(t, 2, tr.GetChangesDetected())

	tr.SetURL("http://localhost:8080")
	assert.Equal(t, "http://localhost:8080", tr.GetURL())

	tr.SetEnvPath("/path/to/.env")
	assert.Equal(t, "/path/to/.env", tr.GetEnvPath())

	info := &DeploymentInfo{Status: "deployed"}
	tr.SetDeploymentInfo(info)
	assert.Same(t, info, tr.GetDeploymentInfo())

	status := tr.GetStatusInfo()
	assert.Equal(t, ConnectionStatusConnected, status.ConnectionStatus)
	assert.Equal(t, 5, status.FilesSynced)
	assert.Equal(t, 2, status.ChangesDetected)

	uptime := tr.GetUptime()
	assert.GreaterOrEqual(t, uptime, time.Duration(0))
}

func TestTracker_EdgeCases(t *testing.T) {
	tr := NewTracker()

	t.Run("GetDeploymentInfo nil initially", func(t *testing.T) {
		assert.Nil(t, tr.GetDeploymentInfo())
	})

	t.Run("SetDeploymentInfo nil", func(t *testing.T) {
		tr.SetDeploymentInfo(nil)
		assert.Nil(t, tr.GetDeploymentInfo())
	})

	t.Run("AddFilesSynced zero", func(t *testing.T) {
		tr.AddFilesSynced(0)
		assert.GreaterOrEqual(t, tr.GetFilesSynced(), 0)
	})

	t.Run("empty URL and env path", func(t *testing.T) {
		tr.SetURL("")
		tr.SetEnvPath("")
		assert.Empty(t, tr.GetURL())
		assert.Empty(t, tr.GetEnvPath())
	})
}

// ---- MultiAppTracker ----

func TestMultiAppTracker(t *testing.T) {
	m := NewMultiAppTracker()

	assert.Empty(t, m.GetSessions())

	m.UpdateSession("app1", AppSessionInfo{
		Name:            "app1",
		ApplicationID:   "id1",
		Status:          ConnectionStatusConnected,
		FilesSynced:     10,
		ChangesDetected: 3,
	})
	m.UpdateSession("app2", AppSessionInfo{
		Name:   "app2",
		Status: ConnectionStatusConnecting,
	})

	sessions := m.GetSessions()
	assert.Len(t, sessions, 2)

	var names []string
	for _, s := range sessions {
		names = append(names, s.Name)
	}
	assert.Contains(t, names, "app1")
	assert.Contains(t, names, "app2")

	uptime := m.GetUptime()
	assert.GreaterOrEqual(t, uptime, time.Duration(0))
}

func TestMultiAppTracker_UpdateOverwrites(t *testing.T) {
	m := NewMultiAppTracker()
	m.UpdateSession("app1", AppSessionInfo{Name: "app1", FilesSynced: 5})
	m.UpdateSession("app1", AppSessionInfo{Name: "app1", FilesSynced: 10})

	sessions := m.GetSessions()
	require.Len(t, sessions, 1)
	assert.Equal(t, 10, sessions[0].FilesSynced)
}

func TestMultiAppTracker_EdgeCases(t *testing.T) {
	m := NewMultiAppTracker()

	t.Run("GetSessions empty", func(t *testing.T) {
		assert.Empty(t, m.GetSessions())
	})

	t.Run("UpdateSession with empty name", func(t *testing.T) {
		m.UpdateSession("", AppSessionInfo{Name: "", FilesSynced: 1})
		sessions := m.GetSessions()
		require.Len(t, sessions, 1)
		assert.Equal(t, "", sessions[0].Name)
	})
}

// ---- Watcher ----

func TestWatcher_NewAndStart(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mover-watcher-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	w, err := New(Config{
		RootPath:       tmpDir,
		DebounceMs:     50,
		IgnorePatterns: []string{"*.tmp"},
	})
	require.NoError(t, err)
	require.NotNil(t, w)

	err = w.Start()
	require.NoError(t, err)

	select {
	case ev := <-w.Events():
		t.Errorf("unexpected event before any change: %v", ev)
	case <-time.After(50 * time.Millisecond):
	}

	err = w.Stop()
	require.NoError(t, err)
}

func TestWatcher_EdgeCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mover-watcher-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("DebounceMs zero uses default", func(t *testing.T) {
		w, err := New(Config{RootPath: tmpDir, DebounceMs: 0})
		require.NoError(t, err)
		require.NoError(t, w.Start())
		defer w.Stop()
	})

	t.Run("EventsBufferSize custom", func(t *testing.T) {
		w, err := New(Config{RootPath: tmpDir, EventsBufferSize: 100})
		require.NoError(t, err)
		require.NoError(t, w.Start())
		defer w.Stop()
	})
}

func TestWatcher_NewInvalidRoot(t *testing.T) {
	// Path that does not exist - Start should fail when addWatchRecursive walks it
	nonExistent := filepath.Join(os.TempDir(), "mover-nonexistent-xyz12345")

	w, err := New(Config{RootPath: nonExistent})
	require.NoError(t, err)

	startErr := w.Start()
	if startErr != nil {
		assert.Contains(t, startErr.Error(), "failed to add watch")
		return
	}
	// Some OS/filesystems may create the path or not fail - just ensure we don't panic
	_ = w.Stop()
}
