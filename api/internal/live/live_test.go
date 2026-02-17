package live

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- FileReceiver ----

func TestFileReceiver_AddChunk(t *testing.T) {
	r := NewFileReceiver("a.go", 3, "abc", "/tmp")

	require.NoError(t, r.AddChunk(0, []byte("chunk0")))
	require.NoError(t, r.AddChunk(1, []byte("chunk1")))
	require.NoError(t, r.AddChunk(2, []byte("chunk2")))

	assert.True(t, r.IsComplete())
	content, err := r.Reassemble()
	require.NoError(t, err)
	assert.Equal(t, []byte("chunk0chunk1chunk2"), content)
}

func TestFileReceiver_AddChunk_TotalChunksZero(t *testing.T) {
	r := NewFileReceiver("a.go", 0, "x", "/tmp")
	err := r.AddChunk(0, []byte("x"))
	assert.Error(t, err)
}

func TestFileReceiver_AddChunk_OutOfBounds(t *testing.T) {
	r := NewFileReceiver("a.go", 2, "x", "/tmp")

	err := r.AddChunk(-1, []byte("x"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of bounds")

	err = r.AddChunk(2, []byte("x"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "out of bounds")

	err = r.AddChunk(3, []byte("x"))
	assert.Error(t, err)
}

func TestFileReceiver_Reassemble_Incomplete(t *testing.T) {
	r := NewFileReceiver("a.go", 3, "x", "/tmp")
	r.AddChunk(0, []byte("a"))
	r.AddChunk(2, []byte("c"))

	_, err := r.Reassemble()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "incomplete")
}

func TestFileReceiver_VerifyChecksum(t *testing.T) {
	content := []byte("hello")
	h := sha256.Sum256(content)
	checksum := hex.EncodeToString(h[:])

	r := NewFileReceiver("a.go", 1, checksum, "/tmp")
	r.AddChunk(0, content)

	assert.True(t, r.VerifyChecksum(content))
	assert.False(t, r.VerifyChecksum([]byte("wrong")))
}

func TestFileReceiver_Reset(t *testing.T) {
	r := NewFileReceiver("a.go", 2, "h1", "/tmp")
	r.AddChunk(0, []byte("a"))
	r.AddChunk(1, []byte("b"))

	r.Reset(1, "h2")
	assert.False(t, r.IsComplete())
	assert.Equal(t, 1, r.TotalChunks)
	assert.Equal(t, "h2", r.Checksum)

	r.AddChunk(0, []byte("x"))
	assert.True(t, r.IsComplete())
	content, _ := r.Reassemble()
	assert.Equal(t, []byte("x"), content)
}

func TestFileReceiver_UpdateMetadata(t *testing.T) {
	r := NewFileReceiver("a.go", 2, "h1", "/tmp")
	r.AddChunk(0, []byte("a"))

	r.UpdateMetadata(2, "h1")
	assert.Equal(t, 2, r.TotalChunks)
	assert.Equal(t, 1, len(r.Chunks))

	r.UpdateMetadata(2, "h2")
	assert.Empty(t, r.Chunks)
	assert.Equal(t, "h2", r.Checksum)
}

func TestFileReceiver_WriteToStaging_Local(t *testing.T) {
	stagingRoot := filepath.Join(os.TempDir(), "nixopus-staging")
	tmpDir := filepath.Join(stagingRoot, "live-test-writestaging")
	require.NoError(t, os.MkdirAll(tmpDir, 0755))
	defer os.RemoveAll(stagingRoot)

	content := []byte("reassembled content")
	h := sha256.Sum256(content)
	checksum := hex.EncodeToString(h[:])

	r := NewFileReceiver("dir/file.go", 1, checksum, tmpDir)
	r.AddChunk(0, content)

	err := r.WriteToStaging(context.Background())
	require.NoError(t, err)

	written, err := os.ReadFile(filepath.Join(tmpDir, "dir", "file.go"))
	require.NoError(t, err)
	assert.Equal(t, content, written)
}

func TestFileReceiver_WriteToStaging_ChecksumMismatch(t *testing.T) {
	stagingRoot := filepath.Join(os.TempDir(), "nixopus-staging")
	tmpDir := filepath.Join(stagingRoot, "live-test-csmismatch")
	require.NoError(t, os.MkdirAll(tmpDir, 0755))
	defer os.RemoveAll(stagingRoot)

	r := NewFileReceiver("a.go", 1, "wrong", tmpDir)
	r.AddChunk(0, []byte("content"))

	err := r.WriteToStaging(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestFileReceiver_ConcurrentAddChunk(t *testing.T) {
	const chunks = 50
	r := NewFileReceiver("big.go", chunks, "x", "/tmp")

	content := []byte("chunk-data")
	var wg sync.WaitGroup
	for i := 0; i < chunks; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			err := r.AddChunk(idx, content)
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()

	assert.True(t, r.IsComplete())
	data, err := r.Reassemble()
	require.NoError(t, err)
	assert.Len(t, data, chunks*len(content))
}

// ---- sanitizePath (via WriteContentToStaging with local path) ----

func TestWriteContentToStaging_Local_Success(t *testing.T) {
	stagingRoot := filepath.Join(os.TempDir(), "nixopus-staging")
	tmpDir := filepath.Join(stagingRoot, "live-test-write")
	require.NoError(t, os.MkdirAll(tmpDir, 0755))
	defer os.RemoveAll(stagingRoot)

	content := []byte("file content")
	h := sha256.Sum256(content)
	checksum := hex.EncodeToString(h[:])

	ctx := context.Background()
	err := WriteContentToStaging(ctx, tmpDir, "src/a.go", content, checksum)
	require.NoError(t, err)

	written, err := os.ReadFile(filepath.Join(tmpDir, "src", "a.go"))
	require.NoError(t, err)
	assert.Equal(t, content, written)
}

func TestWriteContentToStaging_ChecksumMismatch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "live-write-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = WriteContentToStaging(context.Background(), tmpDir, "a.go", []byte("x"), "wrongchecksum")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "checksum mismatch")
}

func TestSanitizePath_EdgeCases(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "live-sanitize-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	t.Run("absolute path rejected", func(t *testing.T) {
		_, err := sanitizePath(tmpDir, "/etc/passwd")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "absolute")
	})

	t.Run("path traversal rejected", func(t *testing.T) {
		_, err := sanitizePath(tmpDir, "../../../etc/passwd")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal")
	})

	t.Run("valid relative path", func(t *testing.T) {
		full, err := sanitizePath(tmpDir, "src/sub/file.go")
		require.NoError(t, err)
		assert.Contains(t, full, "file.go")
		assert.Contains(t, full, tmpDir)
	})

	t.Run("path with dot segments", func(t *testing.T) {
		_, err := sanitizePath(tmpDir, "a/../b/./c")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "..")
	})

	t.Run("leading slash in relative", func(t *testing.T) {
		_, err := sanitizePath(tmpDir, "/relative")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "absolute")
	})
}

// ---- verifyChecksum ----

func TestVerifyChecksum(t *testing.T) {
	content := []byte("test")
	h := sha256.Sum256(content)
	expected := hex.EncodeToString(h[:])

	assert.True(t, verifyChecksum(content, expected))
	assert.False(t, verifyChecksum(content, "wrong"))
	assert.False(t, verifyChecksum([]byte("other"), expected))
}

// ---- Chunker ----

func TestChunkDetectLanguage(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     string
	}{
		{"go", "main.go", "go"},
		{"typescript", "app.tsx", "typescript"},
		{"python", "script.py", "python"},
		{"dockerfile", "Dockerfile", "dockerfile"},
		{"makefile", "Makefile", "makefile"},
		{"unknown", "file.xyz", "unknown"},
		{"empty ext", "Makefile", "makefile"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, chunkDetectLanguage(tt.filename))
		})
	}
}

func TestChunkFile(t *testing.T) {
	t.Run("empty content", func(t *testing.T) {
		chunks := chunkFile("a.go", nil)
		assert.Nil(t, chunks)

		chunks = chunkFile("a.go", []byte("   \n  \n  "))
		assert.Nil(t, chunks)
	})

	t.Run("small file single chunk", func(t *testing.T) {
		content := []byte("package main\n\nfunc main() {}")
		chunks := chunkFile("main.go", content)
		require.Len(t, chunks, 1)
		assert.Equal(t, "main.go", chunks[0].filePath)
	})

	t.Run("go file with boundaries", func(t *testing.T) {
		lines := make([]byte, 0, 200)
		for i := 0; i < 80; i++ {
			lines = append(lines, "func foo"+string(rune('0'+i%10))+"() {}"...)
			lines = append(lines, '\n')
		}
		chunks := chunkFile("pkg.go", lines)
		assert.NotEmpty(t, chunks)
	})
}

func TestIsTextContent(t *testing.T) {
	assert.True(t, isTextContent([]byte("hello\nworld")))
	assert.True(t, isTextContent([]byte("  \t\r\n")))
	assert.False(t, isTextContent([]byte{0}))
	assert.False(t, isTextContent([]byte("hello\x00world")))
	assert.False(t, isTextContent([]byte{1, 2, 3, 4}))
}

// ---- isLocalStagingPath ----

func TestIsLocalStagingPath(t *testing.T) {
	localRoot := filepath.Join(os.TempDir(), "nixopus-staging")
	require.NoError(t, os.MkdirAll(localRoot, 0755))
	defer os.RemoveAll(localRoot)

	assert.True(t, isLocalStagingPath(localRoot))
	assert.True(t, isLocalStagingPath(filepath.Join(localRoot, "user", "env", "app")))
	assert.False(t, isLocalStagingPath("/tmp"))
	assert.False(t, isLocalStagingPath("/var/nixopus/repos"))
}

// ---- Reindex helpers ----

func TestIsSkippedDir(t *testing.T) {
	assert.True(t, isSkippedDir("node_modules"))
	assert.True(t, isSkippedDir(".git"))
	assert.True(t, isSkippedDir(".cache"))
	assert.False(t, isSkippedDir(".github"))
	assert.False(t, isSkippedDir("src"))
	assert.True(t, isSkippedDir(".hidden"))
}

func TestIsBinaryExt(t *testing.T) {
	assert.True(t, isBinaryExt("image.png"))
	assert.True(t, isBinaryExt("file.exe"))
	assert.True(t, isBinaryExt("lib.so"))
	assert.False(t, isBinaryExt("main.go"))
	assert.False(t, isBinaryExt("script.py"))
	assert.False(t, isBinaryExt("README.md"))
}

func TestSliceContains(t *testing.T) {
	sl := []string{"a", "b", "c"}
	assert.True(t, sliceContains(sl, "b"))
	assert.False(t, sliceContains(sl, "x"))
	assert.False(t, sliceContains(nil, "a"))
}

// ---- Service build ----

func TestIsDependencyFile(t *testing.T) {
	assert.True(t, IsDependencyFile("Dockerfile"))
	assert.True(t, IsDependencyFile("dir/Dockerfile"))
	assert.True(t, IsDependencyFile(".dockerignore"))
	assert.False(t, IsDependencyFile("main.go"))
}

func TestMergeEnvVars(t *testing.T) {
	base := map[string]string{"A": "1", "B": "2"}
	override := map[string]string{"B": "overridden", "C": "3"}

	out := mergeEnvVars(base, override)
	assert.Equal(t, "1", out["A"])
	assert.Equal(t, "overridden", out["B"])
	assert.Equal(t, "3", out["C"])
}

func TestMergeEnvVars_EmptyOverride(t *testing.T) {
	base := map[string]string{"A": "1"}
	out := mergeEnvVars(base, nil)
	assert.Equal(t, base, out)

	out = mergeEnvVars(base, map[string]string{})
	assert.Equal(t, base, out)
}

// ---- Manifest persistence ----

func TestNormalizeManifestPath(t *testing.T) {
	assert.Equal(t, "a/b/c", normalizeManifestPath("a/../a/b/./c"))
	assert.Equal(t, "file.go", normalizeManifestPath("./file.go"))
}

// ---- Gateway validateFilePath ----

func TestGateway_ValidateFilePath(t *testing.T) {
	g := NewGateway(nil, nil, nil)
	defer g.Shutdown()

	t.Run("valid relative", func(t *testing.T) {
		assert.NoError(t, g.validateFilePath("src/a.go", "/base"))
	})

	t.Run("absolute rejected", func(t *testing.T) {
		err := g.validateFilePath("/etc/passwd", "/base")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "absolute")
	})

	t.Run("path traversal rejected", func(t *testing.T) {
		err := g.validateFilePath("../../../etc/passwd", "/base")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "path traversal")
	})

	t.Run("empty base path", func(t *testing.T) {
		assert.NoError(t, g.validateFilePath("a/b.go", ""))
	})
}

// ---- StagingManager GetFileReceiver ----

func TestStagingManager_GetFileReceiver(t *testing.T) {
	sm := NewStagingManager(nil, nil)
	appID := uuid.New()
	stagingPath := "/tmp/staging"

	r1 := sm.GetFileReceiver(appID, "a.go", 1, "h1", stagingPath)
	require.NotNil(t, r1)
	assert.Equal(t, "a.go", r1.Path)
	assert.Equal(t, "h1", r1.Checksum)

	r2 := sm.GetFileReceiver(appID, "a.go", 1, "h1", stagingPath)
	assert.Same(t, r1, r2)

	r3 := sm.GetFileReceiver(appID, "a.go", 1, "h2", stagingPath)
	assert.Same(t, r1, r3)
	assert.Equal(t, "h2", r3.Checksum)
}

func TestStagingManager_GetFileReceiver_Concurrent(t *testing.T) {
	sm := NewStagingManager(nil, nil)
	appID := uuid.New()
	stagingPath := "/tmp"

	var wg sync.WaitGroup
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			path := filepath.Join("p", string(rune('a'+idx%26))+".go")
			r := sm.GetFileReceiver(appID, path, 1, "h", stagingPath)
			assert.NotNil(t, r)
		}(i)
	}
	wg.Wait()
}

func TestStagingManager_GetFileReceiver_DifferentPaths(t *testing.T) {
	sm := NewStagingManager(nil, nil)
	appID := uuid.New()

	r1 := sm.GetFileReceiver(appID, "a.go", 1, "h", "/tmp")
	r2 := sm.GetFileReceiver(appID, "b.go", 1, "h", "/tmp")
	assert.NotSame(t, r1, r2)
	assert.Equal(t, "a.go", r1.Path)
	assert.Equal(t, "b.go", r2.Path)
}

func TestStagingManager_RemoveFileReceiver(t *testing.T) {
	sm := NewStagingManager(nil, nil)
	appID := uuid.New()

	r := sm.GetFileReceiver(appID, "a.go", 1, "h", "/tmp")
	require.NotNil(t, r)

	sm.RemoveFileReceiver(appID, "a.go")
	r2 := sm.GetFileReceiver(appID, "a.go", 1, "h", "/tmp")
	assert.NotSame(t, r, r2)
}

// ---- DeleteFileFromStaging (local) ----

func TestDeleteFileFromStaging_Local(t *testing.T) {
	stagingRoot := filepath.Join(os.TempDir(), "nixopus-staging")
	tmpDir := filepath.Join(stagingRoot, "live-test-delete")
	require.NoError(t, os.MkdirAll(tmpDir, 0755))
	defer os.RemoveAll(stagingRoot)

	fpath := filepath.Join(tmpDir, "file.txt")
	require.NoError(t, os.WriteFile(fpath, []byte("x"), 0644))

	ctx := context.Background()
	err := DeleteFileFromStaging(ctx, tmpDir, "file.txt")
	require.NoError(t, err)

	_, err = os.Stat(fpath)
	assert.True(t, os.IsNotExist(err))
}

func TestDeleteFileFromStaging_InvalidPath(t *testing.T) {
	stagingRoot := filepath.Join(os.TempDir(), "nixopus-staging")
	tmpDir := filepath.Join(stagingRoot, "live-test-delete-inv")
	require.NoError(t, os.MkdirAll(tmpDir, 0755))
	defer os.RemoveAll(stagingRoot)

	err := DeleteFileFromStaging(context.Background(), tmpDir, "../../../etc/passwd")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")
}

func TestDeleteFileFromStaging_NotFound(t *testing.T) {
	stagingRoot := filepath.Join(os.TempDir(), "nixopus-staging")
	tmpDir := filepath.Join(stagingRoot, "live-test-delete-nf")
	require.NoError(t, os.MkdirAll(tmpDir, 0755))
	defer os.RemoveAll(stagingRoot)

	err := DeleteFileFromStaging(context.Background(), tmpDir, "nonexistent.txt")
	require.NoError(t, err)
}
