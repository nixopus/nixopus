package utils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/sftp"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newInMemSFTPClient creates an SFTP client connected to an in-memory server.
// Useful for tests that need a real *sftp.Client without SSH.
func newInMemSFTPClient(tb testing.TB) *sftp.Client {
	tb.Helper()
	c1, c2 := netPipe(tb)
	server, err := sftp.NewServer(c1)
	require.NoError(tb, err)
	go server.Serve()
	client, err := sftp.NewClientPipe(c2, c2)
	require.NoError(tb, err)
	return client
}

func netPipe(tb testing.TB) (io.ReadWriteCloser, io.ReadWriteCloser) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(tb, err)
	defer l.Close()

	done := make(chan struct{}, 1)
	done <- struct{}{}

	type result struct {
		conn net.Conn
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		conn, err := l.Accept()
		ch <- result{conn, err}
		if _, ok := <-done; ok {
			l.Close()
			close(done)
		}
	}()

	c1, err := net.Dial("tcp", l.Addr().String())
	require.NoError(tb, err)

	r := <-ch
	require.NoError(tb, r.err)
	return c1, r.conn
}

// ---- Context and Org ID Edge Cases ----

func TestWithSFTPClientFromPool_ContextEdgeCases(t *testing.T) {
	t.Run("missing organization ID", func(t *testing.T) {
		ctx := context.Background()
		err := WithSFTPClientFromPool(ctx, func(*sftp.Client) error { return nil })
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "organization ID required")
	})

	t.Run("empty organization ID", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "")
		err := WithSFTPClientFromPool(ctx, func(*sftp.Client) error { return nil })
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "empty organization ID")
	})

	t.Run("invalid organization ID type", func(t *testing.T) {
		ctx := context.WithValue(context.Background(), types.OrganizationIDKey, 12345)
		err := WithSFTPClientFromPool(ctx, func(*sftp.Client) error { return nil })
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid organization ID type")
	})

	t.Run("organization ID as uuid.UUID", func(t *testing.T) {
		orgID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
		pool := NewSFTPPool(5*time.Minute, func(oid string, _ *ssh.SSHManager) (*sftp.Client, error) {
			assert.Equal(t, orgID.String(), oid)
			return newInMemSFTPClient(t), nil
		})
		sshMgr := ssh.NewSSHManager()
		ctx := context.WithValue(context.Background(), types.OrganizationIDKey, orgID)
		ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
		ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)

		var called bool
		err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error {
			called = true
			assert.NotNil(t, c)
			return nil
		})
		require.NoError(t, err)
		assert.True(t, called)
	})
}

// ---- Concurrency ----

func TestSFTPPool_ConcurrentSameOrg(t *testing.T) {
	const goroutines = 50
	const opsPerGoroutine = 20

	clientCount := int64(0)
	pool := NewSFTPPool(5*time.Minute, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		atomic.AddInt64(&clientCount, 1)
		return newInMemSFTPClient(t), nil
	})
	sshMgr := ssh.NewSSHManager()
	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "org-concurrent")
	ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
	ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)

	// Prime the pool with one client first so concurrent ops reuse it
	err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error { return nil })
	require.NoError(t, err)
	assert.Equal(t, int64(1), atomic.LoadInt64(&clientCount), "prime creates one client")

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error {
					_, err := c.Getwd()
					return err
				})
				assert.NoError(t, err)
			}
		}()
	}
	wg.Wait()

	// Should still have only 1 client (reused across all goroutines)
	assert.Equal(t, int64(1), atomic.LoadInt64(&clientCount),
		"concurrent same-org should reuse single client")
}

func TestSFTPPool_ConcurrentDifferentOrgs(t *testing.T) {
	const numOrgs = 10
	const opsPerOrg = 15

	clientsCreated := sync.Map{}
	pool := NewSFTPPool(5*time.Minute, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		client := newInMemSFTPClient(t)
		clientsCreated.Store(orgID, client)
		return client, nil
	})
	sshMgr := ssh.NewSSHManager()
	ctxBase := context.WithValue(context.Background(), sftpPoolContextKey, pool)
	ctxBase = context.WithValue(ctxBase, sshManagerContextKey, sshMgr)

	var wg sync.WaitGroup
	for i := 0; i < numOrgs; i++ {
		orgID := fmt.Sprintf("org-%d", i)
		ctx := context.WithValue(ctxBase, types.OrganizationIDKey, orgID)
		for j := 0; j < opsPerOrg; j++ {
			wg.Add(1)
			go func(ctx context.Context, oid string) {
				defer wg.Done()
				err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error {
					_, err := c.Getwd()
					return err
				})
				assert.NoError(t, err)
			}(ctx, orgID)
		}
	}
	wg.Wait()

	var count int
	clientsCreated.Range(func(k, v interface{}) bool {
		count++
		c := v.(*sftp.Client)
		c.Close()
		return true
	})
	assert.Equal(t, numOrgs, count, "each org should have its own client")
}

func TestSFTPPool_ConcurrentGetOrCreateRace(t *testing.T) {
	createDelay := make(chan struct{})
	createCount := int64(0)
	pool := NewSFTPPool(5*time.Minute, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		n := atomic.AddInt64(&createCount, 1)
		if n == 1 {
			<-createDelay
		}
		return newInMemSFTPClient(t), nil
	})
	sshMgr := ssh.NewSSHManager()
	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "org-race")
	ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
	ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error { return nil })
			assert.NoError(t, err)
		}()
	}
	time.Sleep(20 * time.Millisecond)
	close(createDelay)
	wg.Wait()

	// When multiple goroutines race on empty pool, some may create clients; losers close theirs.
	// At least one creation; no more than goroutine count. All ops must succeed.
	assert.GreaterOrEqual(t, atomic.LoadInt64(&createCount), int64(1))
	assert.LessOrEqual(t, atomic.LoadInt64(&createCount), int64(5))
}

// ---- Stale Connection Eviction ----

func TestSFTPPool_StaleConnectionEviction(t *testing.T) {
	idleTimeout := 25 * time.Millisecond
	createCount := int64(0)
	pool := NewSFTPPool(idleTimeout, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		atomic.AddInt64(&createCount, 1)
		return newInMemSFTPClient(t), nil
	})
	sshMgr := ssh.NewSSHManager()
	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "org-stale")
	ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
	ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)

	err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error { return nil })
	require.NoError(t, err)
	assert.Equal(t, int64(1), atomic.LoadInt64(&createCount))

	time.Sleep(idleTimeout + 10*time.Millisecond)

	err = WithSFTPClientFromPool(ctx, func(c *sftp.Client) error { return nil })
	require.NoError(t, err)
	assert.Equal(t, int64(2), atomic.LoadInt64(&createCount),
		"stale client should be evicted and new one created")
}

// ---- Closed Connection / Retry ----

func TestSFTPPool_ClosedConnectionRetriesAndEvicts(t *testing.T) {
	attempt := int64(-1)
	pool := NewSFTPPool(5*time.Minute, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		n := atomic.AddInt64(&attempt, 1)
		if n == 0 {
			client := newInMemSFTPClient(t)
			client.Close()
			return client, nil
		}
		return newInMemSFTPClient(t), nil
	})
	sshMgr := ssh.NewSSHManager()
	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "org-closed")
	ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
	ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)

	err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error {
		_, err := c.Getwd()
		return err
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, atomic.LoadInt64(&attempt), int64(1),
		"should have retried after closed connection in fn")
}

func TestSFTPPool_ClosedConnectionErrorFromFnTriggersRetry(t *testing.T) {
	callCount := int64(0)
	pool := NewSFTPPool(5*time.Minute, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		atomic.AddInt64(&callCount, 1)
		return newInMemSFTPClient(t), nil
	})
	sshMgr := ssh.NewSSHManager()
	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "org-fn-retry")
	ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
	ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)

	closedErr := errors.New("use of closed network connection")
	firstCall := true
	err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error {
		if firstCall {
			firstCall = false
			return closedErr
		}
		return nil
	})
	require.NoError(t, err)
	assert.GreaterOrEqual(t, atomic.LoadInt64(&callCount), int64(2),
		"should create new client after fn returns closed connection error")
}

func TestSFTPPool_FactoryReturnsClosedConnectionErrorRetries(t *testing.T) {
	callCount := int64(0)
	pool := NewSFTPPool(5*time.Minute, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		n := atomic.AddInt64(&callCount, 1)
		if n == 1 {
			return nil, errors.New("use of closed network connection")
		}
		return newInMemSFTPClient(t), nil
	})
	sshMgr := ssh.NewSSHManager()
	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "org-factory-retry")
	ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
	ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)

	err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error { return nil })
	require.NoError(t, err)
	assert.GreaterOrEqual(t, atomic.LoadInt64(&callCount), int64(2))
}

// ---- Error Handling Edge Cases ----

func TestSFTPPool_FactoryReturnsPersistentError(t *testing.T) {
	persistentErr := errors.New("SSH connect: dial tcp: connection refused")
	pool := NewSFTPPool(5*time.Minute, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		return nil, persistentErr
	})
	sshMgr := ssh.NewSSHManager()
	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "org-err")
	ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
	ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)

	err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error { return nil })
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

func TestSFTPPool_NonRetryableErrorFromFn(t *testing.T) {
	pool := NewSFTPPool(5*time.Minute, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		return newInMemSFTPClient(t), nil
	})
	sshMgr := ssh.NewSSHManager()
	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "org-nonretry")
	ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
	ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)

	bizErr := errors.New("file not found")
	err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error {
		return bizErr
	})
	assert.Error(t, err)
	assert.Equal(t, bizErr, err)
}

func TestSFTPPool_ExhaustsRetriesOnClosedConnection(t *testing.T) {
	pool := NewSFTPPool(5*time.Minute, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		return nil, errors.New("use of closed network connection")
	})
	sshMgr := ssh.NewSSHManager()
	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "org-exhaust")
	ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
	ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)

	err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error { return nil })
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "org-exhaust")
	assert.Contains(t, err.Error(), "use of closed network connection")
}

// ---- Touch Updates LastUsed ----

func TestSFTPPool_TouchPreventsEviction(t *testing.T) {
	idleTimeout := 50 * time.Millisecond
	createCount := int64(0)
	pool := NewSFTPPool(idleTimeout, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		atomic.AddInt64(&createCount, 1)
		return newInMemSFTPClient(t), nil
	})
	sshMgr := ssh.NewSSHManager()
	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "org-touch")
	ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
	ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)

	for i := 0; i < 5; i++ {
		err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error { return nil })
		require.NoError(t, err)
		time.Sleep(idleTimeout / 3)
	}
	assert.Equal(t, int64(1), atomic.LoadInt64(&createCount),
		"touch on success should keep client alive")
}

// ---- Evict Only Exact Match ----

func TestSFTPPool_EvictOnlyExactClient(t *testing.T) {
	pool := NewSFTPPool(5*time.Minute, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		return newInMemSFTPClient(t), nil
	})
	sshMgr := ssh.NewSSHManager()
	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "org-evict")
	ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
	ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)

	var storedClient *sftp.Client
	err := WithSFTPClientFromPool(ctx, func(c *sftp.Client) error {
		storedClient = c
		return nil
	})
	require.NoError(t, err)

	otherClient := newInMemSFTPClient(t)
	defer otherClient.Close()
	pool.evict("org-evict", otherClient)

	err = WithSFTPClientFromPool(ctx, func(c *sftp.Client) error {
		assert.Same(t, storedClient, c, "evict with wrong client should not remove pool entry")
		return nil
	})
	require.NoError(t, err)
}

// ---- sftp_utils (ReadFile, ReadFileBytes, FileExists, FilesExist, WalkRemote, ReadFileBytesFromClient) ----

func testCtxWithPool(t *testing.T) (context.Context, *SFTPPool, func()) {
	t.Helper()
	pool := NewSFTPPool(5*time.Minute, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		return newInMemSFTPClient(t), nil
	})
	sshMgr := ssh.NewSSHManager()
	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "org-sftp-utils")
	ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
	ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)
	return ctx, pool, func() {}
}

func TestReadFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sftptest-readfile-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	content := "hello world\nline two"
	filePath := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	ctx, _, cleanup := testCtxWithPool(t)
	defer cleanup()

	got, err := ReadFile(ctx, filePath)
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestReadFile_NotFound(t *testing.T) {
	ctx, _, cleanup := testCtxWithPool(t)
	defer cleanup()

	_, err := ReadFile(ctx, "/nonexistent/path/file.txt")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestReadFileBytes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sftptest-readfilebytes-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	content := []byte("binary \x00 data")
	filePath := filepath.Join(tmpDir, "bin.dat")
	require.NoError(t, os.WriteFile(filePath, content, 0644))

	ctx, _, cleanup := testCtxWithPool(t)
	defer cleanup()

	got, err := ReadFileBytes(ctx, filePath)
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestReadFileBytesFromClient(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sftptest-readfilebytesclient-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	content := []byte("from client")
	filePath := filepath.Join(tmpDir, "client.txt")
	require.NoError(t, os.WriteFile(filePath, content, 0644))

	ctx, _, cleanup := testCtxWithPool(t)
	defer cleanup()

	var got []byte
	err = WithSFTPClient(ctx, func(c *sftp.Client) error {
		var err error
		got, err = ReadFileBytesFromClient(c, filePath)
		return err
	})
	require.NoError(t, err)
	assert.Equal(t, content, got)
}

func TestReadFileBytesFromClient_NotFound(t *testing.T) {
	ctx, _, cleanup := testCtxWithPool(t)
	defer cleanup()

	err := WithSFTPClient(ctx, func(c *sftp.Client) error {
		_, err := ReadFileBytesFromClient(c, "/nonexistent/file.txt")
		return err
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open file")
}

func TestFileExists(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sftptest-fileexists-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath := filepath.Join(tmpDir, "exists.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("x"), 0644))

	ctx, _, cleanup := testCtxWithPool(t)
	defer cleanup()

	assert.True(t, FileExists(ctx, filePath))
	assert.False(t, FileExists(ctx, filepath.Join(tmpDir, "missing.txt")))
}

func TestFilesExist(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sftptest-filesexist-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	pathA := filepath.Join(tmpDir, "a.txt")
	pathB := filepath.Join(tmpDir, "b.txt")
	pathC := filepath.Join(tmpDir, "c.txt")
	require.NoError(t, os.WriteFile(pathA, []byte("a"), 0644))
	require.NoError(t, os.WriteFile(pathC, []byte("c"), 0644))

	ctx, _, cleanup := testCtxWithPool(t)
	defer cleanup()

	result := FilesExist(ctx, []string{pathA, pathB, pathC})
	assert.True(t, result[pathA])
	assert.False(t, result[pathB])
	assert.True(t, result[pathC])
}

func TestFilesExist_OnErrorMarksAllNonExistent(t *testing.T) {
	pool := NewSFTPPool(5*time.Minute, func(orgID string, _ *ssh.SSHManager) (*sftp.Client, error) {
		return nil, errors.New("connection refused")
	})
	sshMgr := ssh.NewSSHManager()
	ctx := context.WithValue(context.Background(), types.OrganizationIDKey, "org-fail")
	ctx = context.WithValue(ctx, sftpPoolContextKey, pool)
	ctx = context.WithValue(ctx, sshManagerContextKey, sshMgr)

	paths := []string{"/a", "/b"}
	result := FilesExist(ctx, paths)
	assert.False(t, result["/a"])
	assert.False(t, result["/b"])
}

func TestWalkRemote(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sftptest-walk-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "f1.txt"), []byte("1"), 0644))
	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "sub"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "sub", "f2.txt"), []byte("2"), 0644))

	ctx, _, cleanup := testCtxWithPool(t)
	defer cleanup()

	var visited []string
	err = WalkRemote(ctx, tmpDir, func(client *sftp.Client, path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info != nil {
			visited = append(visited, path)
		}
		return nil
	})
	require.NoError(t, err)
	assert.Contains(t, visited, tmpDir)
	assert.Contains(t, visited, filepath.Join(tmpDir, "f1.txt"))
	assert.Contains(t, visited, filepath.Join(tmpDir, "sub"))
	assert.Contains(t, visited, filepath.Join(tmpDir, "sub", "f2.txt"))
}

func TestWalkRemote_SkipDir(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "sftptest-walkskip-")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	require.NoError(t, os.MkdirAll(filepath.Join(tmpDir, "skipme"), 0755))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "skipme", "inner.txt"), []byte("x"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "keep.txt"), []byte("keep"), 0644))

	ctx, _, cleanup := testCtxWithPool(t)
	defer cleanup()

	var visited []string
	err = WalkRemote(ctx, tmpDir, func(client *sftp.Client, path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info != nil && info.IsDir() && filepath.Base(path) == "skipme" {
			return filepath.SkipDir
		}
		if info != nil {
			visited = append(visited, path)
		}
		return nil
	})
	require.NoError(t, err)
	assert.Contains(t, visited, filepath.Join(tmpDir, "keep.txt"))
	assert.NotContains(t, visited, filepath.Join(tmpDir, "skipme", "inner.txt"))
}

func TestWithSFTPClient(t *testing.T) {
	ctx, _, cleanup := testCtxWithPool(t)
	defer cleanup()

	var called bool
	err := WithSFTPClient(ctx, func(c *sftp.Client) error {
		called = true
		_, err := c.Getwd()
		return err
	})
	require.NoError(t, err)
	assert.True(t, called)
}
