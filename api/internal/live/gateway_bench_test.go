package live

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
)

// BenchmarkFileReceiverAddChunk measures chunk insertion throughput under concurrency.
// Simulates multiple goroutines adding chunks to the same receiver (worst-case lock contention).
func BenchmarkFileReceiverAddChunk(b *testing.B) {
	totalChunks := 100
	chunkSize := 64 * 1024
	h := sha256.Sum256([]byte("test"))
	checksum := hex.EncodeToString(h[:])
	receiver := NewFileReceiver("test/file.go", totalChunks, checksum, "/tmp/staging")

	data := make([]byte, chunkSize)
	for i := range data {
		data[i] = byte(i % 256)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = receiver.AddChunk(i%totalChunks, data)
	}
}

// BenchmarkFileReceiverAddChunkParallel measures throughput when many goroutines add chunks.
// Worst-case: all workers contend on the same receiver.
func BenchmarkFileReceiverAddChunkParallel(b *testing.B) {
	totalChunks := 100
	chunkSize := 1024
	h := sha256.Sum256([]byte("test"))
	checksum := hex.EncodeToString(h[:])
	data := make([]byte, chunkSize)

	for _, p := range []int{1, 4, 8, 16, 32} {
		p := p
		b.Run(fmt.Sprintf("workers=%d", p), func(b *testing.B) {
			receiver := NewFileReceiver("test/file.go", totalChunks, checksum, "/tmp/staging")
			b.ResetTimer()
			b.ReportAllocs()
			b.SetParallelism(p)
			b.RunParallel(func(pb *testing.PB) {
				i := 0
				for pb.Next() {
					_ = receiver.AddChunk(i%totalChunks, data)
					i++
				}
			})
		})
	}
}

// BenchmarkFileReceiverManyPaths measures throughput when each goroutine has its own receiver.
// Best-case: no lock contention, simulates many files syncing in parallel.
func BenchmarkFileReceiverManyPaths(b *testing.B) {
	totalChunks := 100
	chunkSize := 1024
	h := sha256.Sum256([]byte("test"))
	checksum := hex.EncodeToString(h[:])
	data := make([]byte, chunkSize)
	numReceivers := 256
	receivers := make([]*FileReceiver, numReceivers)
	for i := 0; i < numReceivers; i++ {
		receivers[i] = NewFileReceiver(fmt.Sprintf("file_%d.go", i), totalChunks, checksum, "/tmp/staging")
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			r := receivers[i%numReceivers]
			_ = r.AddChunk(i%totalChunks, data)
			i++
		}
	})
}

// BenchmarkStagingManagerGetFileReceiver measures GetFileReceiver under concurrent load.
// Uses a real StagingManager with nil dependencies (GetFileReceiver doesn't use them).
func BenchmarkStagingManagerGetFileReceiver(b *testing.B) {
	sm := &StagingManager{
		fileReceivers: make(map[uuid.UUID]map[string]*FileReceiver),
	}
	appID := uuid.New()
	stagingPath := "/tmp/staging"

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		path := "file_" + string(rune('a'+i%26)) + ".go"
		_ = sm.GetFileReceiver(appID, path, 10, "abc123", stagingPath)
	}
}

// BenchmarkStagingManagerGetFileReceiverParallel measures scalability across goroutines.
func BenchmarkStagingManagerGetFileReceiverParallel(b *testing.B) {
	sm := &StagingManager{
		fileReceivers: make(map[uuid.UUID]map[string]*FileReceiver),
	}
	appID := uuid.New()
	stagingPath := "/tmp/staging"

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			path := "file_" + string(rune('a'+i%26)) + ".go"
			_ = sm.GetFileReceiver(appID, path, 10, "abc123", stagingPath)
			i++
		}
	})
}

// BenchmarkStagingManagerMultiApp measures GetFileReceiver under multi-tenant load.
// Many appIDs × many paths simulates production (multiple devs, multiple files each).
func BenchmarkStagingManagerMultiApp(b *testing.B) {
	sm := &StagingManager{
		fileReceivers: make(map[uuid.UUID]map[string]*FileReceiver),
	}
	stagingPath := "/tmp/staging"
	numApps := 64
	numPaths := 32
	appIDs := make([]uuid.UUID, numApps)
	for i := 0; i < numApps; i++ {
		appIDs[i] = uuid.New()
	}

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			appID := appIDs[i%numApps]
			path := fmt.Sprintf("src/file_%d.go", i%numPaths)
			_ = sm.GetFileReceiver(appID, path, 10, "abc123", stagingPath)
			i++
		}
	})
}

// BenchmarkGatewaySessionEnvStore measures concurrent GetSessionEnv and mutations.
func BenchmarkGatewaySessionEnvStore(b *testing.B) {
	g := &Gateway{
		sessionEnvStore: make(map[string]map[string]string),
	}
	appID := uuid.New()
	g.sessionEnvStore[appID.String()] = map[string]string{"KEY": "value", "FOO": "bar"}

	b.Run("GetSessionEnv", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = g.GetSessionEnv(appID)
		}
	})

	b.Run("GetSessionEnvParallel", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				_ = g.GetSessionEnv(appID)
			}
		})
	})

	b.Run("SessionEnvWriteParallel", func(b *testing.B) {
		b.ResetTimer()
		b.ReportAllocs()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				id := uuid.New()
				vars := map[string]string{"K": "v", "F": "b"}
				g.sessionEnvStoreMu.Lock()
				g.sessionEnvStore[id.String()] = vars
				g.sessionEnvStoreMu.Unlock()
				g.sessionEnvStoreMu.Lock()
				delete(g.sessionEnvStore, id.String())
				g.sessionEnvStoreMu.Unlock()
				i++
			}
		})
	})
}

// BenchmarkFileChunkPipeline measures the in-memory flow for one full file sync.
// GetFileReceiver → AddChunk (all chunks) → IsComplete → Reassemble → RemoveFileReceiver.
// No I/O; simulates handleFileContent's chunk handling for a 100-chunk file.
func BenchmarkFileChunkPipeline(b *testing.B) {
	sm := &StagingManager{fileReceivers: make(map[uuid.UUID]map[string]*FileReceiver)}
	appID := uuid.New()
	stagingPath := "/tmp/staging"
	totalChunks := 100
	chunkSize := 1024
	h := sha256.Sum256([]byte("test"))
	checksum := hex.EncodeToString(h[:])
	data := make([]byte, chunkSize)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		path := fmt.Sprintf("file_%d.go", i%64)
		receiver := sm.GetFileReceiver(appID, path, totalChunks, checksum, stagingPath)
		for c := 0; c < totalChunks; c++ {
			_ = receiver.AddChunk(c, data)
		}
		_, _ = receiver.Reassemble()
		sm.RemoveFileReceiver(appID, path)
	}
}

// BenchmarkCompletionJobChannel measures job enqueue throughput with single consumer.
func BenchmarkCompletionJobChannel(b *testing.B) {
	jobs := make(chan *fileCompletionJob, completionBuffer())
	job := makeFileCompletionJob()

	done := make(chan struct{})
	go func() {
		for range jobs {
		}
		close(done)
	}()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		jobs <- job
	}
	close(jobs)
	<-done
}

// BenchmarkCompletionJobChannelMultiWorker measures throughput with 16 workers (matches production).
// Many producers enqueue; 16 consumers drain. Tests real gateway topology.
func BenchmarkCompletionJobChannelMultiWorker(b *testing.B) {
	const numWorkers = 16
	jobs := make(chan *fileCompletionJob, completionBuffer())
	job := makeFileCompletionJob()

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
			}
		}()
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		jobs <- job
	}
	close(jobs)
	wg.Wait()
}

func makeFileCompletionJob() *fileCompletionJob {
	appID := uuid.New()
	orgID := uuid.New()
	return &fileCompletionJob{
		content:       make([]byte, 1024),
		path:          "test.go",
		checksum:      "abc123",
		stagingPath:   "/tmp/staging",
		appCtx:        &ApplicationContext{ApplicationID: appID, OrganizationID: orgID, StagingPath: "/tmp/staging"},
		applicationID: appID,
	}
}

// BenchmarkActiveConnsLookup measures activeConns map lookup under RLock.
// Simulates the hot path for sendPipelineProgress, sendBuildStatus, etc.
func BenchmarkActiveConnsLookup(b *testing.B) {
	g := &Gateway{activeConns: make(map[uuid.UUID]*activeConn)}
	appID := uuid.New()
	g.activeConnsMu.Lock()
	g.activeConns[appID] = &activeConn{} // nil conn/handler for benchmark
	g.activeConnsMu.Unlock()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		g.activeConnsMu.RLock()
		_ = g.activeConns[appID]
		g.activeConnsMu.RUnlock()
	}
}

// BenchmarkActiveConnsLookupParallel measures scalability of activeConns under concurrent reads.
func BenchmarkActiveConnsLookupParallel(b *testing.B) {
	g := &Gateway{activeConns: make(map[uuid.UUID]*activeConn)}
	ids := make([]uuid.UUID, 64)
	g.activeConnsMu.Lock()
	for i := 0; i < 64; i++ {
		id := uuid.New()
		ids[i] = id
		g.activeConns[id] = &activeConn{}
	}
	g.activeConnsMu.Unlock()

	b.ResetTimer()
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			g.activeConnsMu.RLock()
			_ = g.activeConns[ids[i%len(ids)]]
			g.activeConnsMu.RUnlock()
			i++
		}
	})
}
