package mover

import (
	"encoding/json"
	"fmt"
	"testing"
)

// BenchmarkBuildFromPaths measures Merkle tree construction from path→checksum map.
func BenchmarkBuildFromPaths(b *testing.B) {
	sizes := []int{10, 100, 500, 1000, 5000}
	for _, n := range sizes {
		n := n
		leaves := make(map[string]string, n)
		for i := 0; i < n; i++ {
			leaves[fmt.Sprintf("src/file_%d.go", i)] = "a1b2c3d4e5f6" + fmt.Sprintf("%02d", i%100)
		}
		b.Run(fmt.Sprintf("paths=%d", n), func(b *testing.B) {
			b.ResetTimer()
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = BuildFromPaths(leaves)
			}
		})
	}
}

// BenchmarkDiffAgainst measures Merkle diff between local and server leaves.
func BenchmarkDiffAgainst(b *testing.B) {
	local := make(map[string]string, 500)
	server := make(map[string]string, 500)
	for i := 0; i < 500; i++ {
		p := fmt.Sprintf("src/file_%d.go", i)
		local[p] = "hash" + fmt.Sprintf("%d", i)
		server[p] = "hash" + fmt.Sprintf("%d", i)
	}
	// 10% diff
	for i := 0; i < 50; i++ {
		local[fmt.Sprintf("src/changed_%d.go", i)] = "newhash"
	}
	delete(server, "src/file_0.go")

	tree := BuildFromPaths(local)
	serverNorm := make(map[string]string, len(server))
	for p, c := range server {
		serverNorm[normalizeMerklePath(p)] = c
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, _ = DiffAgainst(tree, serverNorm)
	}
}

// BenchmarkComputeSimhash measures simhash computation for codebase fingerprint.
func BenchmarkComputeSimhash(b *testing.B) {
	leaves := make(map[string]string, 200)
	for i := 0; i < 200; i++ {
		leaves[fmt.Sprintf("file_%d.go", i)] = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = ComputeSimhash(leaves)
	}
}

// BenchmarkParallelSyncJobPool measures job struct pool Get/Put overhead.
func BenchmarkParallelSyncJobPool(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		j := parallelSyncJobPool.Get().(*parallelSyncJob)
		j.index = i
		j.file = "test.go"
		parallelSyncJobPool.Put(j)
	}
}

// BenchmarkParseEnvelope measures recvEnvelope unmarshal (no double encode).
func BenchmarkParseEnvelope(b *testing.B) {
	msg := []byte(`{"type":"manifest","timestamp":"2024-01-01T00:00:00Z","payload":{"paths":{"a":"x","b":"y"},"root_hash":"abc"}}`)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var env recvEnvelope
		_ = json.Unmarshal(msg, &env)
	}
}

// BenchmarkSyncStateSetPath measures SetPath under lock (hot path during sync).
func BenchmarkSyncStateSetPath(b *testing.B) {
	s := NewSyncState("/tmp/bench-sync-state.json")
	appID := "app-123"
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		s.SetPath(appID, fmt.Sprintf("src/file_%d.go", i%100), "checksum")
	}
}

// BenchmarkSyncStateGetPaths measures GetPaths under RLock.
func BenchmarkSyncStateGetPaths(b *testing.B) {
	s := NewSyncState("/tmp/bench-sync-state.json")
	appID := "app-123"
	for i := 0; i < 100; i++ {
		s.SetPath(appID, fmt.Sprintf("src/file_%d.go", i), "hash")
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = s.GetPaths(appID)
	}
}

// BenchmarkTrackerGetStatusInfo measures hot path for status rendering.
func BenchmarkTrackerGetStatusInfo(b *testing.B) {
	t := NewTracker()
	t.SetConnectionStatus(ConnectionStatusConnected)
	t.SetServiceStatus(ServiceStatusRunning)
	t.AddFilesSynced(100)
	t.AddFilesSynced(50)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = t.GetStatusInfo()
	}
}

// BenchmarkBatchGitCheckIgnoreInput prep measures building the path list for batch input.
// Simulates the string building we do when piping to git check-ignore --stdin.
func BenchmarkBatchGitCheckIgnoreInput(b *testing.B) {
	paths := make([]string, 2000)
	for i := 0; i < 2000; i++ {
		paths[i] = fmt.Sprintf("src/pkg/module/file_%d.go", i)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf := make([]byte, 0, 2000*50)
		for _, p := range paths {
			buf = append(buf, p...)
			buf = append(buf, '\n')
		}
		_ = buf
	}
}
