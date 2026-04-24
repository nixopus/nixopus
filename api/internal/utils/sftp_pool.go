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

// sftpPoolMaxAttemptsOverride defaults to -1 (use maxRetries). Tests may set 0 to exercise
// the post-loop return path, or another value to cap attempts.
var sftpPoolMaxAttemptsOverride = -1

func effectiveSftpPoolMaxAttempts() int {
	if sftpPoolMaxAttemptsOverride >= 0 {
		return sftpPoolMaxAttemptsOverride
	}
	return maxRetries
}

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
// When nil, the pool uses sshMgr.Borrow and newSftpFromGophClient (real SSH + SFTP subsystem).
type SFTPClientFactory func(orgID string, sshMgr *ssh.SSHManager) (*sftp.Client, error)

// sshGetManagerFromContext defaults to ssh.GetSSHManagerFromContext; tests may point it at a stub
// to cover sftpPoolSSHManager without full app init.
var sshGetManagerFromContext = ssh.GetSSHManagerFromContext

// SFTPPool provides org-scoped SFTP client reuse to avoid connection churn.
// Clients are cached per organization and evicted on idle timeout or connection errors.
type SFTPPool struct {
	mu            sync.RWMutex
	clients       map[string]*pooledSFTP
	idleTimeout   time.Duration
	clientFactory SFTPClientFactory // when non-nil, used instead of the real SSH/SFTP path
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
	orgID, err := sftpPoolOrganizationID(ctx)
	if err != nil {
		return err
	}
	pool := sftpPoolFromContext(ctx)
	sshMgr, err := sftpPoolSSHManager(ctx)
	if err != nil {
		return err
	}
	cacheKey := sftpPoolCacheKey(ctx, orgID)

	for attempt := 0; attempt < effectiveSftpPoolMaxAttempts(); attempt++ {
		client, release, fromPool, createErr := pool.getOrCreate(ctx, cacheKey, sshMgr)
		if client == nil {
			if sftpPooledGetShouldRetryCreate(createErr, attempt) {
				continue
			}
			return fmt.Errorf("failed to get SFTP client for org %s: %w", orgID, createErr)
		}

		var once sync.Once
		doRelease := func() { once.Do(release) }

		err = fn(client)
		if err == nil {
			pool.touch(cacheKey)
			doRelease()
			return nil
		}

		doRelease()
		if !isClosedConnectionError(err) {
			return err
		}

		pool.evict(cacheKey, client)
		if fromPool {
			sshMgr.CloseConnection("")
		}
		if !sftpPooledFnShouldRetryAfterClosed(attempt) {
			return err
		}
	}
	// Compiler cannot prove loop always returns (continue + retry paths).
	return fmt.Errorf("internal error: SFTP pool retry loop exited without result")
}

func sftpPoolOrganizationID(ctx context.Context) (string, error) {
	orgIDAny := ctx.Value(types.OrganizationIDKey)
	if orgIDAny == nil {
		return "", fmt.Errorf("organization ID required for SFTP pool")
	}
	switch v := orgIDAny.(type) {
	case string:
		if v == "" {
			return "", fmt.Errorf("empty organization ID")
		}
		return v, nil
	case uuid.UUID:
		return v.String(), nil
	default:
		return "", fmt.Errorf("invalid organization ID type: %T", orgIDAny)
	}
}

func sftpPoolFromContext(ctx context.Context) *SFTPPool {
	p := ctx.Value(sftpPoolContextKey)
	if p == nil {
		return globalSFTPPool
	}
	if pp, ok := p.(*SFTPPool); ok {
		return pp
	}
	return globalSFTPPool
}

func sftpPoolSSHManager(ctx context.Context) (*ssh.SSHManager, error) {
	mgr, err := sshGetManagerFromContext(ctx)
	if err == nil {
		return mgr, nil
	}
	m := ctx.Value(sshManagerContextKey)
	if m == nil {
		return nil, err
	}
	mm, ok := m.(*ssh.SSHManager)
	if !ok {
		return nil, err
	}
	return mm, nil
}

func sftpPoolCacheKey(ctx context.Context, orgID string) string {
	if serverIDStr, ok := ctx.Value(types.ServerIDKey).(string); ok && serverIDStr != "" {
		return orgID + ":" + serverIDStr
	}
	return orgID
}

func sftpPooledGetShouldRetryCreate(createErr error, attempt int) bool {
	n := effectiveSftpPoolMaxAttempts()
	return createErr != nil && isClosedConnectionError(createErr) && attempt < n-1
}

func sftpPooledFnShouldRetryAfterClosed(attempt int) bool {
	return attempt < effectiveSftpPoolMaxAttempts()-1
}

// getOrCreate returns (client, release, fromPool, error).
// Caller must call release() when done (e.g. via defer) to avoid eviction races.
func (p *SFTPPool) getOrCreate(ctx context.Context, orgID string, sshMgr *ssh.SSHManager) (*sftp.Client, func(), bool, error) {
	noop := func() {}

	p.mu.Lock()
	if entry, ok := p.clients[orgID]; ok {
		if reuse, client, rel := tryAcquirePooledClient(entry, p.idleTimeout); reuse {
			p.mu.Unlock()
			return client, rel, true, nil
		}
		// Stale, idle client: evict under lock, then build a new one without the lock.
		entry.client.Close()
		delete(p.clients, orgID)
	}
	p.mu.Unlock()

	sftpClient, sshRelease, err := p.openNewPooledSftpClient(orgID, sshMgr)
	if err != nil {
		return nil, noop, false, err
	}

	return p.insertOrUseConcurrentPooledClient(orgID, sftpClient, sshRelease)
}

// tryAcquirePooledClient returns (ok, client, release) when a cached client is still valid.
// If !ok, the entry is idle and should be evicted (caller holds the pool lock).
func tryAcquirePooledClient(entry *pooledSFTP, idleTimeout time.Duration) (ok bool, client *sftp.Client, release func()) {
	if entry.inUse.Load() > 0 {
		return acquirePooledClientRef(entry)
	}
	if time.Since(entry.lastUsed) <= idleTimeout {
		return acquirePooledClientRef(entry)
	}
	return false, nil, nil
}

func acquirePooledClientRef(entry *pooledSFTP) (ok bool, client *sftp.Client, release func()) {
	entry.inUse.Add(1)
	return true, entry.client, func() { entry.inUse.Add(-1) }
}

func (p *SFTPPool) openNewPooledSftpClient(orgID string, sshMgr *ssh.SSHManager) (*sftp.Client, func(), error) {
	if p.clientFactory != nil {
		c, err := p.clientFactory(orgID, sshMgr)
		if err != nil {
			return nil, nil, err
		}
		return c, nil, nil
	}

	sshClient, release, err := sshMgr.Borrow("")
	if err != nil {
		return nil, nil, fmt.Errorf("SSH connect: %w", err)
	}
	sftpClient, err := newSftpFromGophClient(sshClient)
	if err != nil {
		release()
		if isClosedConnectionError(err) {
			sshMgr.CloseConnection("")
		}
		return nil, nil, fmt.Errorf("SFTP subsystem: %w", err)
	}
	return sftpClient, release, nil
}

func (p *SFTPPool) insertOrUseConcurrentPooledClient(orgID string, sftpClient *sftp.Client, sshRelease func()) (*sftp.Client, func(), bool, error) {
	p.mu.Lock()
	if existing, ok := p.clients[orgID]; ok {
		// Another goroutine added one while we were creating.
		p.mu.Unlock()
		if sshRelease != nil {
			sshRelease()
		}
		_ = sftpClient.Close()
		existing.inUse.Add(1)
		r := func() { existing.inUse.Add(-1) }
		return existing.client, r, true, nil
	}

	entry := &pooledSFTP{
		client:     sftpClient,
		lastUsed:   time.Now(),
		sshRelease: sshRelease, // nil for clientFactory path
	}
	entry.inUse.Store(1)
	p.clients[orgID] = entry
	p.mu.Unlock()
	r := func() { entry.inUse.Add(-1) }
	return sftpClient, r, false, nil
}

func (p *SFTPPool) evict(orgID string, c *sftp.Client) {
	p.mu.Lock()
	defer p.mu.Unlock()
	e, ok := p.clients[orgID]
	if !ok || e.client != c {
		return
	}
	p.releaseAndClosePooledEntry(e, orgID)
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
		p.releaseAndClosePooledEntry(entry, orgID)
	}
}

func (p *SFTPPool) releaseAndClosePooledEntry(e *pooledSFTP, orgID string) {
	if e.sshRelease != nil {
		e.sshRelease()
	}
	_ = e.client.Close()
	delete(p.clients, orgID)
}

// InvalidateSFTPPoolForOrg removes the cached SFTP client for an organization
// from the global pool. Call when the org's SSH config changes.
func InvalidateSFTPPoolForOrg(orgID string) {
	globalSFTPPool.EvictOrg(orgID)
}
