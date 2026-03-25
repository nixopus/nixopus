package utils

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/pkg/sftp"
)

func init() {
	ssh.RegisterInvalidateHook(func(orgID uuid.UUID) {
		InvalidateSFTPPoolForOrg(orgID.String())
	})
}

const (
	defaultSFTPIdleTimeout = 5 * time.Minute
)

// Context keys for test injection (used by sftp_pool_test.go)
type sftpPoolContextKeyType struct{}
type sshManagerContextKeyType struct{}

var (
	sftpPoolContextKey   = sftpPoolContextKeyType{}
	sshManagerContextKey = sshManagerContextKeyType{}
)

type pooledSFTP struct {
	client     *sftp.Client
	lastUsed   time.Time
	inUse      atomic.Int64 // active callers using this client; eviction only when 0
	sshRelease func()       // called when evicting to release the underlying SSH connection
}

// SFTPClientFactory creates SFTP clients. Used for dependency injection in tests.
// When nil, the pool uses sshMgr.Connect() and sshClient.NewSftp().
type SFTPClientFactory func(orgID string, sshMgr *ssh.SSHManager) (*sftp.Client, error)

// SFTPPool provides org-scoped SFTP client reuse to avoid connection churn.
// Clients are cached per organization and evicted on idle timeout or connection errors.
type SFTPPool struct {
	mu            sync.RWMutex
	clients       map[string]*pooledSFTP
	idleTimeout   time.Duration
	clientFactory SFTPClientFactory // when non-nil, used instead of sshMgr for creating clients (for tests)
}

var globalSFTPPool = &SFTPPool{
	clients:     make(map[string]*pooledSFTP),
	idleTimeout: defaultSFTPIdleTimeout,
}

// NewSFTPPool creates a new pool with the given idle timeout.
// If factory is non-nil, it is used to create clients instead of the real SSH flow (for testing).
func NewSFTPPool(idleTimeout time.Duration, factory SFTPClientFactory) *SFTPPool {
	return &SFTPPool{
		clients:       make(map[string]*pooledSFTP),
		idleTimeout:   idleTimeout,
		clientFactory: factory,
	}
}

// WithSFTPClientFromPool runs fn with an SFTP client from the org-scoped pool.
// Context must have types.OrganizationIDKey set. Falls back to local staging (no SFTP) is not applicable.
// Evicts stale clients on connection errors; creates new client when pool empty or evicted.
// For testing: use context.WithValue(ctx, sftpPoolContextKey, pool) and sshManagerContextKey for overrides.
func WithSFTPClientFromPool(ctx context.Context, fn func(*sftp.Client) error) error {
	orgIDAny := ctx.Value(types.OrganizationIDKey)
	if orgIDAny == nil {
		return fmt.Errorf("organization ID required for SFTP pool")
	}
	var orgID string
	switch v := orgIDAny.(type) {
	case string:
		orgID = v
	case uuid.UUID:
		orgID = v.String()
	default:
		return fmt.Errorf("invalid organization ID type: %T", orgIDAny)
	}
	if orgID == "" {
		return fmt.Errorf("empty organization ID")
	}

	pool := globalSFTPPool
	if p := ctx.Value(sftpPoolContextKey); p != nil {
		if pp, ok := p.(*SFTPPool); ok {
			pool = pp
		}
	}

	sshMgr, err := ssh.GetSSHManagerFromContext(ctx)
	if err != nil {
		if m := ctx.Value(sshManagerContextKey); m != nil {
			if mm, ok := m.(*ssh.SSHManager); ok {
				sshMgr = mm
				err = nil
			}
		}
	}
	if err != nil {
		return err
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		client, release, fromPool, createErr := pool.getOrCreate(ctx, orgID, sshMgr)
		if client == nil {
			if createErr != nil && isClosedConnectionError(createErr) && attempt < maxRetries-1 {
				// Stale connection (e.g. "use of closed network connection"); evicted by getOrCreate, retry
				continue
			}
			if createErr != nil {
				return fmt.Errorf("failed to get SFTP client for org %s: %w", orgID, createErr)
			}
			return fmt.Errorf("failed to get SFTP client for org %s (unknown error)", orgID)
		}
		var releaseOnce sync.Once
		doRelease := func() { releaseOnce.Do(release) }
		defer doRelease()

		err = fn(client)
		if err != nil {
			if isClosedConnectionError(err) {
				doRelease() // Release before evict so refcount is accurate
				pool.evict(orgID, client)
				if fromPool {
					sshMgr.CloseConnection("")
				}
				if attempt < maxRetries-1 {
					continue
				}
			}
			return err
		}
		pool.touch(orgID)
		return nil
	}
	return fmt.Errorf("SFTP operation failed after %d attempts", maxRetries)
}

// getOrCreate returns (client, release, fromPool, error).
// Caller must call release() when done (e.g. via defer) to avoid eviction races.
func (p *SFTPPool) getOrCreate(ctx context.Context, orgID string, sshMgr *ssh.SSHManager) (*sftp.Client, func(), bool, error) {
	noop := func() {}

	p.mu.Lock()
	if entry, ok := p.clients[orgID]; ok {
		// Only evict when idle AND no one is using it; otherwise reuse
		inUse := entry.inUse.Load()
		idle := time.Since(entry.lastUsed) > p.idleTimeout
		if inUse > 0 || !idle {
			entry.inUse.Add(1)
			client := entry.client
			release := func() { entry.inUse.Add(-1) }
			p.mu.Unlock()
			return client, release, true, nil
		}
		entry.client.Close()
		delete(p.clients, orgID)
	}
	p.mu.Unlock()

	// Create new client outside lock to avoid blocking other orgs
	var sftpClient *sftp.Client
	var sshRelease func() // non-nil when created via Borrow
	if p.clientFactory != nil {
		var err error
		sftpClient, err = p.clientFactory(orgID, sshMgr)
		if err != nil {
			return nil, noop, false, err
		}
	} else {
		sshClient, release, err := sshMgr.Borrow("")
		if err != nil {
			return nil, noop, false, fmt.Errorf("SSH connect: %w", err)
		}
		sshRelease = release
		sftpClient, err = sshClient.NewSftp()
		if err != nil {
			sshRelease()
			if isClosedConnectionError(err) {
				sshMgr.CloseConnection("")
			}
			return nil, noop, false, fmt.Errorf("SFTP subsystem: %w", err)
		}
	}

	p.mu.Lock()
	if existing, ok := p.clients[orgID]; ok {
		// Another goroutine added one while we were creating
		p.mu.Unlock()
		if sshRelease != nil {
			sshRelease() // not using our new client
		}
		sftpClient.Close()
		existing.inUse.Add(1)
		release := func() { existing.inUse.Add(-1) }
		return existing.client, release, true, nil
	}
	entry := &pooledSFTP{
		client:     sftpClient,
		lastUsed:   time.Now(),
		sshRelease: sshRelease, // nil for clientFactory path
	}
	entry.inUse.Store(1)
	p.clients[orgID] = entry
	p.mu.Unlock()
	release := func() { entry.inUse.Add(-1) }
	return sftpClient, release, false, nil
}

func (p *SFTPPool) evict(orgID string, c *sftp.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if entry, ok := p.clients[orgID]; ok && entry.client == c {
		if entry.sshRelease != nil {
			entry.sshRelease()
		}
		entry.client.Close()
		delete(p.clients, orgID)
	}
}

func (p *SFTPPool) touch(orgID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if entry, ok := p.clients[orgID]; ok {
		entry.lastUsed = time.Now()
	}
}

// EvictOrg removes and closes the SFTP client for a specific organization.
// Safe to call even if no client is cached for the org.
func (p *SFTPPool) EvictOrg(orgID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if entry, ok := p.clients[orgID]; ok {
		if entry.sshRelease != nil {
			entry.sshRelease()
		}
		entry.client.Close()
		delete(p.clients, orgID)
	}
}

// InvalidateSFTPPoolForOrg removes the cached SFTP client for an organization
// from the global pool. Call when the org's SSH config changes.
func InvalidateSFTPPoolForOrg(orgID string) {
	globalSFTPPool.EvictOrg(orgID)
}
