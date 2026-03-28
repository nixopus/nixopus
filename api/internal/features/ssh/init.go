package ssh

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/melbahja/goph"
	"github.com/nixopus/nixopus/api/internal/config"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh/service"
	sshstorage "github.com/nixopus/nixopus/api/internal/features/ssh/storage"
	"github.com/nixopus/nixopus/api/internal/types"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

const (
	dialTimeout         = 15 * time.Second
	KeepaliveInterval   = 15 * time.Second
	KeepaliveMaxMissed  = 4
	keepaliveReqTimeout = 10 * time.Second
)

// SSH represents a single SSH connection configuration
type SSH struct {
	PrivateKey          string `json:"private_key"`
	PublicKey           string `json:"public_key"`
	Host                string `json:"host"`
	ProxyHost           string `json:"proxy_host"`
	User                string `json:"user"`
	Port                uint   `json:"port"`
	Password            string `json:"password"`
	PrivateKeyProtected string `json:"private_key_protected"`
}

// connectionPoolEntry represents a pooled SSH connection with metadata
type connectionPoolEntry struct {
	client        *goph.Client
	lastUsed      time.Time
	inUse         atomic.Int64 // active borrowers; cleanup only closes when 0
	mu            sync.RWMutex
	stopKeepalive chan struct{} // closed to stop the keepalive goroutine for this connection
}

// SSHConnectFunc creates SSH connections. Used for dependency injection in tests.
// When non-nil, ConnectWithID uses it instead of sshClient.ConnectWithRetry.
type SSHConnectFunc func(id string) (*goph.Client, error)

// SSHManager manages multiple SSH clients and provides a unified interface
// For now, it defaults to single client mode for backward compatibility
// In the future, it can be extended to support multiple clients/discoveries
type SSHManager struct {
	clients        map[string]*SSH                 // Map of client ID to SSH config
	defaultID      string                          // ID of the default client
	mu             sync.RWMutex                    // Mutex for thread safe access
	pool           map[string]*connectionPoolEntry // Connection pool by client ID
	poolMu         sync.RWMutex                    // Mutex for connection pool
	connectingMu   map[string]*sync.Mutex          // Mutex per client ID to prevent concurrent connection creation
	connectingMuMu sync.Mutex                      // Mutex for accessing connectingMu map
	maxIdleTime    time.Duration                   // Maximum idle time before closing connection
	connectFunc    SSHConnectFunc                  // when non-nil, used for ConnectWithID (for tests)
	done           chan struct{}                   // closed to stop the cleanup goroutine
}

var (
	// serverManagers caches SSHManager per server (ssh_key.id).
	// orgToServerIDs is the reverse index: orgID → []serverID for org-level eviction.
	// No orgDefaultServer cache — default-server lookup always goes to DB (~0.1ms index
	// scan), which is correct under horizontal scaling and direct DB writes by other services.
	serverManagers   = make(map[string]*SSHManager)
	orgToServerIDs   = make(map[string][]string)
	serverManagersMu sync.RWMutex

	// onInvalidateHooks are called (in order) after an org's SSH manager is
	// evicted. Downstream packages (docker, SFTP pool, caddy) register hooks
	// via RegisterInvalidateHook so they can flush their own caches without
	// creating import cycles.
	onInvalidateHooks   []func(orgID uuid.UUID)
	onInvalidateHooksMu sync.Mutex
)

// RegisterInvalidateHook adds a callback that fires whenever an org's SSH
// cache is invalidated. Use this from init() in packages that maintain their
// own SSH-derived caches (Docker service, SFTP pool, Caddy tunnels).
func RegisterInvalidateHook(fn func(orgID uuid.UUID)) {
	onInvalidateHooksMu.Lock()
	onInvalidateHooks = append(onInvalidateHooks, fn)
	onInvalidateHooksMu.Unlock()
}

func fireInvalidateHooks(orgID uuid.UUID) {
	onInvalidateHooksMu.Lock()
	hooks := make([]func(uuid.UUID), len(onInvalidateHooks))
	copy(hooks, onInvalidateHooks)
	onInvalidateHooksMu.Unlock()
	for _, fn := range hooks {
		fn(orgID)
	}
}

// GetSSHManagerForOrganization returns an SSHManager for the org's default server.
// Always queries the DB for the current default (cheap partial-unique-index scan, ~0.1ms).
// This is intentionally not cached: it keeps behavior correct across multiple API
// instances and when other services write SSH keys directly to the database.
// Signature unchanged — all existing callers continue to work.
func GetSSHManagerForOrganization(ctx context.Context, orgID uuid.UUID) (*SSHManager, error) {
	if config.GlobalStore == nil {
		return nil, fmt.Errorf("global store not initialized, ensure config.Init() has been called")
	}

	sshKeyStorage := sshstorage.SSHKeyStorage{DB: config.GlobalStore.DB, Ctx: ctx}
	defaultKey, err := sshKeyStorage.GetDefaultSSHKeyByOrganizationID(orgID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("no default server configured for organization %s", orgID.String())
		}
		return nil, fmt.Errorf("failed to get default SSH key for organization %s: %w", orgID.String(), err)
	}

	return GetSSHManagerForServer(ctx, orgID, defaultKey.ID)
}

// GetSSHManagerForServer returns an SSHManager for a specific server (ssh_key.id).
// Validates that serverID belongs to orgID. Caches the manager under serverID.
// This is the single place where managers are built and stored.
func GetSSHManagerForServer(ctx context.Context, orgID uuid.UUID, serverID uuid.UUID) (*SSHManager, error) {
	if config.GlobalStore == nil {
		return nil, fmt.Errorf("global store not initialized")
	}

	if orgID == uuid.Nil {
		return nil, fmt.Errorf("orgID must not be nil")
	}
	if serverID == uuid.Nil {
		return nil, fmt.Errorf("serverID must not be nil")
	}

	serverIDStr := serverID.String()

	// Fast path: cache hit (read lock only)
	serverManagersMu.RLock()
	if mgr, exists := serverManagers[serverIDStr]; exists {
		serverManagersMu.RUnlock()
		return mgr, nil
	}
	serverManagersMu.RUnlock()

	// Slow path: load from DB under write lock (double-checked)
	serverManagersMu.Lock()
	defer serverManagersMu.Unlock()

	if mgr, exists := serverManagers[serverIDStr]; exists {
		return mgr, nil
	}

	// Validate server belongs to org
	sshKeyStorage := sshstorage.SSHKeyStorage{DB: config.GlobalStore.DB, Ctx: ctx}
	sshKey, err := sshKeyStorage.GetSSHKeyByID(serverID)
	if err != nil {
		return nil, fmt.Errorf("server %s not found: %w", serverIDStr, err)
	}
	if sshKey.OrganizationID != orgID {
		return nil, fmt.Errorf("server %s does not belong to organization %s", serverIDStr, orgID.String())
	}

	// Build SSH config from the specific key
	sshSvc := service.NewSSHKeyService(config.GlobalStore, ctx, logger.NewLogger())
	sshConfig, err := sshSvc.GetSSHConfigForKey(sshKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build SSH config for server %s: %w", serverIDStr, err)
	}

	sshClient := NewSSHFromConfig(sshConfig)
	if sshClient == nil {
		return nil, fmt.Errorf("SSH config is nil for server %s", serverIDStr)
	}
	if len(sshClient.PrivateKey) == 0 && len(sshClient.Password) == 0 {
		return nil, fmt.Errorf("SSH config for server %s has no credentials", serverIDStr)
	}
	if len(sshClient.PrivateKey) > 0 && !strings.HasPrefix(sshClient.PrivateKey, "-----BEGIN") {
		return nil, fmt.Errorf("SSH private key for server %s is not a valid PEM key", serverIDStr)
	}

	manager := NewSSHManager()
	if err := manager.AddClient("default", sshClient); err != nil {
		return nil, fmt.Errorf("failed to register SSH client for server %s: %w", serverIDStr, err)
	}

	serverManagers[serverIDStr] = manager
	orgToServerIDs[orgID.String()] = appendUnique(orgToServerIDs[orgID.String()], serverIDStr)

	return manager, nil
}

// appendUnique appends s to slice only if not already present.
func appendUnique(slice []string, s string) []string {
	for _, v := range slice {
		if v == s {
			return slice
		}
	}
	return append(slice, s)
}

// InvalidateSSHManagerCache evicts all cached SSH managers for an organization.
// Closes all pooled connections, clears the reverse index, then fires registered hooks.
// Safe to call with uuid.Nil (no-op).
func InvalidateSSHManagerCache(orgID uuid.UUID) {
	if orgID == uuid.Nil {
		return
	}
	orgIDStr := orgID.String()

	serverManagersMu.Lock()
	serverIDs := orgToServerIDs[orgIDStr]
	var toClose []*SSHManager
	for _, sid := range serverIDs {
		if mgr, exists := serverManagers[sid]; exists {
			toClose = append(toClose, mgr)
			delete(serverManagers, sid)
		}
	}
	delete(orgToServerIDs, orgIDStr)
	serverManagersMu.Unlock()

	for _, mgr := range toClose {
		mgr.Close()
	}

	fireInvalidateHooks(orgID)
}

// InvalidateServerManagerCache evicts a single server's cached SSH manager.
// Removes it from the reverse index. Safe to call with uuid.Nil (no-op).
func InvalidateServerManagerCache(serverID uuid.UUID) {
	if serverID == uuid.Nil {
		return
	}
	serverIDStr := serverID.String()

	serverManagersMu.Lock()
	mgr, exists := serverManagers[serverIDStr]
	if !exists {
		serverManagersMu.Unlock()
		return
	}
	delete(serverManagers, serverIDStr)

	var foundOrgID string
	for orgIDStr, ids := range orgToServerIDs {
		for i, sid := range ids {
			if sid == serverIDStr {
				foundOrgID = orgIDStr
				orgToServerIDs[orgIDStr] = append(ids[:i], ids[i+1:]...)
				break
			}
		}
	}
	serverManagersMu.Unlock()

	mgr.Close()
	if foundOrgID != "" {
		if parsed, err := uuid.Parse(foundOrgID); err == nil {
			fireInvalidateHooks(parsed)
		}
	}
}

// InvalidateAllSSHManagerCaches clears every cached manager. Useful at shutdown
// or when a global config change (e.g. key rotation) affects all orgs.
func InvalidateAllSSHManagerCaches() {
	serverManagersMu.Lock()
	snapshot := serverManagers
	serverManagers = make(map[string]*SSHManager)
	orgToServerIDs = make(map[string][]string)
	serverManagersMu.Unlock()

	for _, mgr := range snapshot {
		if mgr != nil {
			mgr.Close()
		}
	}
}

// GetSSHManagerFromContext extracts organization ID from context and returns the appropriate SSHManager.
// This is the new primary entry point for getting SSH managers.
// The organization ID should be set in context by the auth middleware via types.OrganizationIDKey.
// Uses the global store set during config.Init().
func GetSSHManagerFromContext(ctx context.Context) (*SSHManager, error) {
	orgIDAny := ctx.Value(types.OrganizationIDKey)
	if orgIDAny == nil {
		return nil, fmt.Errorf("organization ID not found in context")
	}

	var orgID uuid.UUID
	switch v := orgIDAny.(type) {
	case string:
		var err error
		orgID, err = uuid.Parse(v)
		if err != nil {
			return nil, fmt.Errorf("invalid organization ID in context: %w", err)
		}
	case uuid.UUID:
		orgID = v
	default:
		return nil, fmt.Errorf("unexpected organization ID type in context: %T", v)
	}

	return GetSSHManagerForOrganization(ctx, orgID)
}

// NewSSHManager creates a new empty SSH manager.
// Clients must be added via AddClient() or use GetSSHManagerForOrganization() / GetSSHManagerFromContext()
// to get an organization-specific manager with pre-configured clients.
// For most use cases, prefer GetSSHManagerFromContext() to get an organization-specific manager.
func NewSSHManager() *SSHManager {
	return newSSHManager(5*time.Minute, nil)
}

// NewSSHManagerForTest creates a manager with an optional connection factory for testing.
// When connectFunc is non-nil, ConnectWithID uses it instead of real SSH. maxIdleTime can be 0 for default.
func NewSSHManagerForTest(connectFunc SSHConnectFunc, maxIdleTime time.Duration) *SSHManager {
	if maxIdleTime == 0 {
		maxIdleTime = 5 * time.Minute
	}
	return newSSHManager(maxIdleTime, connectFunc)
}

func newSSHManager(maxIdleTime time.Duration, connectFunc SSHConnectFunc) *SSHManager {
	manager := &SSHManager{
		clients:      make(map[string]*SSH),
		defaultID:    "default",
		pool:         make(map[string]*connectionPoolEntry),
		connectingMu: make(map[string]*sync.Mutex),
		maxIdleTime:  maxIdleTime,
		connectFunc:  connectFunc,
		done:         make(chan struct{}),
	}
	go manager.cleanupIdleConnections()
	return manager
}

// AddClient adds a new SSH client to the manager with a unique ID
func (m *SSHManager) AddClient(id string, sshClient *SSH) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id == "" {
		return fmt.Errorf("client ID cannot be empty")
	}
	if sshClient == nil {
		return fmt.Errorf("SSH client cannot be nil")
	}

	m.clients[id] = sshClient
	return nil
}

// GetClient retrieves an SSH client by ID, or returns the default client if ID is empty
func (m *SSHManager) GetClient(id string) (*SSH, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if id == "" {
		id = m.defaultID
	}

	client, exists := m.clients[id]
	if !exists {
		return nil, fmt.Errorf("SSH client with ID '%s' not found", id)
	}

	return client, nil
}

// SetDefault sets the default client ID
func (m *SSHManager) SetDefault(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.clients[id]; !exists {
		return fmt.Errorf("SSH client with ID '%s' does not exist", id)
	}

	m.defaultID = id
	return nil
}

// Connect connects to the default SSH client with connection pooling
// This maintains backward compatibility with the single-client approach
func (m *SSHManager) Connect() (*goph.Client, error) {
	return m.ConnectWithID("")
}

// Borrow returns a pooled SSH connection and a release function.
// Caller must call release() when done to avoid cleanup closing the connection while in use.
// Prefer Borrow over Connect when holding the connection across async work (e.g. SFTP pool).
func (m *SSHManager) Borrow(id string) (*goph.Client, func(), error) {
	if id == "" {
		id = m.defaultID
	}
	client, err := m.ConnectWithID(id)
	if err != nil {
		return nil, func() {}, err
	}
	// Increment inUse so cleanup won't close this connection until release is called.
	// Small race window: cleanup could run between ConnectWithID return and our increment.
	// If that happens, client may be closed; next use will fail and caller can retry.
	m.poolMu.Lock()
	entry, exists := m.pool[id]
	if exists && entry != nil {
		entry.inUse.Add(1)
	}
	m.poolMu.Unlock()
	if !exists || entry == nil {
		return client, func() {}, nil // no pool entry (e.g. freshly created, inUse lives in entry)
	}
	release := func() { entry.inUse.Add(-1) }
	return client, release, nil
}

// IsNoDefaultServerError returns true when the error means the org has no default server configured.
// Callers should map this to 503 Service Unavailable.
func IsNoDefaultServerError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "no default server configured")
}

// IsClosedConnectionError checks if the error indicates a closed or stale network connection.
// Aligned with SFTP pool detection: EOF, broken pipe, connection reset, etc.
func IsClosedConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if err == io.EOF {
		return true
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "use of closed network connection") ||
		strings.Contains(errMsg, "connection closed") ||
		strings.Contains(errMsg, "connection lost") ||
		strings.Contains(errMsg, "EOF") ||
		strings.Contains(errMsg, "broken pipe") ||
		strings.Contains(errMsg, "connection reset by peer") ||
		strings.Contains(errMsg, "unexpected packet")
}

// isClosedConnectionError is an internal alias for IsClosedConnectionError
func isClosedConnectionError(err error) bool {
	return IsClosedConnectionError(err)
}

// StartKeepalive sends periodic keepalive@openssh.com requests over the SSH
// connection to prevent NAT/firewall idle timeouts and detect dead connections
// early. When maxMissed consecutive keepalives fail, the client is closed so
// callers see immediate errors and trigger reconnection. The goroutine stops
// when stop is closed or maxMissed is exceeded.
func StartKeepalive(client *goph.Client, interval time.Duration, maxMissed int, stop <-chan struct{}) {
	if client == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		missed := 0
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if !sendKeepalive(client, keepaliveReqTimeout) {
					missed++
					if missed >= maxMissed {
						client.Close()
						return
					}
				} else {
					missed = 0
				}
			}
		}
	}()
}

// sendKeepalive sends a single keepalive request with a timeout guard.
// Returns true if the remote replied successfully within the deadline.
// The timeout prevents a hung TCP connection from blocking the keepalive loop.
func sendKeepalive(client *goph.Client, timeout time.Duration) bool {
	done := make(chan bool, 1)
	go func() {
		_, _, err := client.SendRequest("keepalive@openssh.com", true, nil)
		done <- (err == nil)
	}()
	select {
	case ok := <-done:
		return ok
	case <-time.After(timeout):
		return false
	}
}

// isConnectionAlive checks if an SSH connection is still valid by attempting to create a test session
func (m *SSHManager) isConnectionAlive(client *goph.Client) bool {
	if client == nil {
		return false
	}
	// Try to create a new session to validate the connection
	session, err := client.NewSession()
	if err != nil {
		return false
	}
	session.Close()
	return true
}

// NewSessionWithRetry creates a new SSH session with automatic retry on closed connection errors.
// This method handles stale connections by removing them from the pool and retrying.
// The returned session should be closed by the caller when done.
func (m *SSHManager) NewSessionWithRetry(id string) (*ssh.Session, error) {
	const maxRetries = 2

	if id == "" {
		id = m.defaultID
	}

	for attempt := 0; attempt < maxRetries; attempt++ {
		// Get connection from pool (reuses existing or creates new if needed)
		client, err := m.ConnectWithID(id)
		if err != nil {
			return nil, fmt.Errorf("failed to connect via SSH: %w", err)
		}

		session, err := client.NewSession()
		if err != nil {
			if isClosedConnectionError(err) {
				// Remove the bad connection from pool and retry
				m.CloseConnection(id)
				if attempt < maxRetries-1 {
					continue
				}
			}
			return nil, fmt.Errorf("failed to create SSH session: %w", err)
		}

		return session, nil
	}

	return nil, fmt.Errorf("failed to create SSH session after %d attempts due to connection issues", maxRetries)
}

// ConnectWithID connects to a specific SSH client by ID with connection pooling
func (m *SSHManager) ConnectWithID(id string) (*goph.Client, error) {
	if id == "" {
		id = m.defaultID
	}

	// Try to get existing connection from pool
	m.poolMu.RLock()
	entry, exists := m.pool[id]
	var client *goph.Client
	var err error
	if exists && entry != nil {
		entry.mu.RLock()
		client = entry.client
		entry.mu.RUnlock()
	}
	m.poolMu.RUnlock()

	// Validate the pooled connection is still alive
	if client != nil {
		// Check if connection was recently used (within last 30 seconds) - skip validation if so
		// This avoids unnecessary validation checks that can fail due to server rate limiting
		m.poolMu.RLock()
		entry, entryExists := m.pool[id]
		recentlyUsed := false
		var lastUsed time.Time
		if entryExists && entry != nil {
			entry.mu.RLock()
			lastUsed = entry.lastUsed
			entry.mu.RUnlock()
			// If used within last 30 seconds, consider it still valid without validation
			if time.Since(lastUsed) < 30*time.Second {
				recentlyUsed = true
			}
		}
		m.poolMu.RUnlock()

		// Only validate if not recently used (to avoid unnecessary checks that might fail)
		// Recently used connections are assumed to be alive to avoid false negatives
		if recentlyUsed || m.isConnectionAlive(client) {
			// Connection is valid, update last used time and return
			m.poolMu.Lock()
			if entry, exists := m.pool[id]; exists && entry != nil {
				entry.mu.Lock()
				entry.lastUsed = time.Now()
				entry.mu.Unlock()
			}
			m.poolMu.Unlock()
			return client, nil
		}
		// Connection is dead, remove it from pool and stop its keepalive
		m.poolMu.Lock()
		if entry, exists := m.pool[id]; exists {
			entry.mu.Lock()
			if entry.client == client {
				if entry.stopKeepalive != nil {
					close(entry.stopKeepalive)
					entry.stopKeepalive = nil
				}
				entry.client = nil
			}
			entry.mu.Unlock()
			delete(m.pool, id)
		}
		m.poolMu.Unlock()
	}

	// No valid pooled connection available, create new one
	// Use per-client-ID mutex to prevent concurrent connection creation
	m.connectingMuMu.Lock()
	connMu, exists := m.connectingMu[id]
	if !exists {
		connMu = &sync.Mutex{}
		m.connectingMu[id] = connMu
	}
	m.connectingMuMu.Unlock()

	connMu.Lock()
	defer connMu.Unlock()

	// Double-check: another goroutine might have created the connection while we waited
	m.poolMu.RLock()
	entry, exists = m.pool[id]
	var existingClient *goph.Client
	if exists && entry != nil {
		entry.mu.RLock()
		existingClient = entry.client
		entry.mu.RUnlock()
	}
	m.poolMu.RUnlock()

	if existingClient != nil && m.isConnectionAlive(existingClient) {
		// Connection was created by another goroutine, use it
		m.poolMu.Lock()
		if entry, exists := m.pool[id]; exists {
			entry.mu.Lock()
			entry.lastUsed = time.Now()
			entry.mu.Unlock()
		}
		m.poolMu.Unlock()
		return existingClient, nil
	}

	if m.connectFunc != nil {
		client, err = m.connectFunc(id)
	} else {
		var sshClient *SSH
		sshClient, err = m.GetClient(id)
		if err != nil {
			return nil, err
		}
		client, err = sshClient.ConnectWithRetry()
	}
	if err != nil {
		return nil, err
	}
	if client == nil {
		return nil, fmt.Errorf("SSH connection factory returned nil client")
	}

	// Store in pool and start keepalive to prevent NAT/firewall idle timeouts
	stopCh := make(chan struct{})
	StartKeepalive(client, KeepaliveInterval, KeepaliveMaxMissed, stopCh)
	m.poolMu.Lock()
	m.pool[id] = &connectionPoolEntry{
		client:        client,
		lastUsed:      time.Now(),
		stopKeepalive: stopCh,
	}
	m.poolMu.Unlock()

	return client, nil
}

// cleanupIdleConnections periodically closes idle connections.
// Stops when the manager's done channel is closed.
func (m *SSHManager) cleanupIdleConnections() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.done:
			return
		case <-ticker.C:
			now := time.Now()
			m.poolMu.Lock()
			for id, entry := range m.pool {
				entry.mu.Lock()
				inUse := entry.inUse.Load()
				idle := entry.client != nil && now.Sub(entry.lastUsed) > m.maxIdleTime
				if idle && inUse == 0 {
					if entry.stopKeepalive != nil {
						close(entry.stopKeepalive)
						entry.stopKeepalive = nil
					}
					entry.client.Close()
					entry.client = nil
					delete(m.pool, id)
				}
				entry.mu.Unlock()
			}
			m.poolMu.Unlock()
		}
	}
}

// CloseConnection closes a specific connection in the pool
func (m *SSHManager) CloseConnection(id string) {
	if id == "" {
		id = m.defaultID
	}
	m.poolMu.Lock()
	if entry, exists := m.pool[id]; exists {
		entry.mu.Lock()
		if entry.stopKeepalive != nil {
			close(entry.stopKeepalive)
			entry.stopKeepalive = nil
		}
		if entry.client != nil {
			entry.client.Close()
			entry.client = nil
		}
		entry.mu.Unlock()
		delete(m.pool, id)
	}
	m.poolMu.Unlock()
}

// Close shuts down this manager: stops the cleanup goroutine, stops all
// keepalive goroutines, and closes all pooled connections. The manager
// must not be used after calling Close.
func (m *SSHManager) Close() {
	select {
	case <-m.done:
		// already closed
	default:
		close(m.done)
	}

	m.poolMu.Lock()
	for id, entry := range m.pool {
		entry.mu.Lock()
		if entry.stopKeepalive != nil {
			close(entry.stopKeepalive)
			entry.stopKeepalive = nil
		}
		if entry.client != nil {
			entry.client.Close()
			entry.client = nil
		}
		entry.mu.Unlock()
		delete(m.pool, id)
	}
	m.poolMu.Unlock()
}

// RunCommand runs a command on the default SSH client using the connection pool.
func (m *SSHManager) RunCommand(cmd string) (string, error) {
	return m.RunCommandWithID("", cmd)
}

// RunCommandWithID runs a command on a specific SSH client by ID using the
// connection pool. Stale connections are automatically evicted and retried.
func (m *SSHManager) RunCommandWithID(id string, cmd string) (string, error) {
	session, err := m.NewSessionWithRetry(id)
	if err != nil {
		return "", fmt.Errorf("failed to get SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), err
	}
	return string(output), nil
}

// ListClients returns a list of all client IDs
func (m *SSHManager) ListClients() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.clients))
	for id := range m.clients {
		ids = append(ids, id)
	}
	return ids
}

// GetDefaultSSH returns the default SSH client struct
// This is useful when you need direct access to the SSH struct (e.g., for accessing Host field)
func (m *SSHManager) GetDefaultSSH() (*SSH, error) {
	return m.GetClient("")
}

// GetOrganizationSSH returns the organization-specific SSH client struct
// This manager is organization-specific, so this returns the organization's SSH client
func (m *SSHManager) GetOrganizationSSH() (*SSH, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	client, exists := m.clients[m.defaultID]
	if !exists {
		return nil, fmt.Errorf("SSH client not found for organization")
	}
	return client, nil
}

// GetSSHHost returns the SSH host for the organization's SSH client
func (m *SSHManager) GetSSHHost() (string, error) {
	sshClient, err := m.GetOrganizationSSH()
	if err != nil {
		return "", fmt.Errorf("failed to get organization SSH client: %w", err)
	}
	if sshClient.Host == "" {
		return "", fmt.Errorf("SSH host is not configured for organization")
	}
	return sshClient.Host, nil
}

// GetUpstreamHost returns the proxy host for the organization if configured,
// otherwise falls back to the SSH host.
func (m *SSHManager) GetUpstreamHost() (string, error) {
	sshClient, err := m.GetOrganizationSSH()
	if err != nil {
		return "", fmt.Errorf("failed to get organization SSH client: %w", err)
	}
	if sshClient.ProxyHost != "" {
		return sshClient.ProxyHost, nil
	}
	if sshClient.Host == "" {
		return "", fmt.Errorf("neither proxy host nor SSH host is configured for organization")
	}
	return sshClient.Host, nil
}

// GetSSHUser returns the SSH user for the organization's SSH client
func (m *SSHManager) GetSSHUser() (string, error) {
	sshClient, err := m.GetOrganizationSSH()
	if err != nil {
		return "", fmt.Errorf("failed to get organization SSH client: %w", err)
	}
	if sshClient.User == "" {
		return "", fmt.Errorf("SSH user is not configured for organization")
	}
	return sshClient.User, nil
}

// GetSSHConfig returns the SSH config struct for read-only access
// Returns the organization-specific SSH configuration
func (m *SSHManager) GetSSHConfig() (*SSH, error) {
	return m.GetOrganizationSSH()
}

// NewSSHFromConfig creates a new SSH client from a custom SSHConfig
// This is useful when you need to create SSH clients with different configurations
// Example usage for multi-client scenarios:
//
//	client1 := NewSSHFromConfig(&types.SSHConfig{Host: "server1", User: "user1", ...})
//	client2 := NewSSHFromConfig(&types.SSHConfig{Host: "server2", User: "user2", ...})
//	manager := NewSSHManager()
//	manager.AddClient("server1", client1)
//	manager.AddClient("server2", client2)
func NewSSHFromConfig(sshConfig *types.SSHConfig) *SSH {
	if sshConfig == nil {
		return nil // Don't fallback to config - SSH config must be provided
	}
	return &SSH{
		PrivateKey:          sshConfig.PrivateKey,
		Host:                sshConfig.Host,
		ProxyHost:           sshConfig.ProxyHost,
		User:                sshConfig.User,
		Port:                sshConfig.Port,
		Password:            sshConfig.Password,
		PrivateKeyProtected: sshConfig.PrivateKeyProtected,
	}
}

func (s *SSH) ConnectWithPassword() (*goph.Client, error) {
	if s.Password == "" {
		return nil, fmt.Errorf("password is required for SSH connection")
	}

	auth := goph.Password(s.Password)

	client, err := goph.NewConn(&goph.Config{
		User:     s.User,
		Addr:     s.Host,
		Port:     uint(s.Port),
		Auth:     auth,
		Timeout:  dialTimeout,
		Callback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection with password: %w", err)
	}

	return client, nil
}

// Connect creates a new SSH connection (used internally, prefer SSHManager.Connect for pooling)
func (s *SSH) Connect() (*goph.Client, error) {
	return s.ConnectWithRetry()
}

// ConnectWithRetry attempts to connect with exponential backoff
func (s *SSH) ConnectWithRetry() (*goph.Client, error) {
	if s.User == "" || s.Host == "" {
		return nil, fmt.Errorf("user and host are required for SSH connection")
	}

	hasPrivateKey := len(s.PrivateKey) > 0
	hasPassword := len(s.Password) > 0

	var keyErr error
	if hasPrivateKey {
		client, err := s.ConnectWithPrivateKey()
		if err == nil {
			return client, nil
		}
		keyErr = err
	}

	if !hasPassword {
		if keyErr != nil {
			return nil, fmt.Errorf("private key auth failed: %w; no password configured as fallback", keyErr)
		}
		return nil, fmt.Errorf("no authentication method available: private key and password are both empty")
	}

	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt) * baseDelay
			time.Sleep(delay)
		}

		client, err := s.ConnectWithPassword()
		if err == nil {
			return client, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("failed to connect with both private key and password after %d attempts: %w", maxRetries, lastErr)
}

func (s *SSH) ConnectWithPrivateKey() (*goph.Client, error) {
	if s.PrivateKey == "" {
		return nil, fmt.Errorf("private key is required for SSH connection")
	}

	auth, err := goph.RawKey(s.PrivateKey, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH auth from private key: %w", err)
	}

	client, err := goph.NewConn(&goph.Config{
		User:     s.User,
		Addr:     s.Host,
		Port:     uint(s.Port),
		Auth:     auth,
		Timeout:  dialTimeout,
		Callback: ssh.InsecureIgnoreHostKey(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection with private key: %w", err)
	}

	return client, nil
}

// func (s *SSH) ConnectWithPrivateKeyProtected() (*goph.Client, error) {
// 	auth, err := goph.Key(s.PrivateKeyProtected, "")

// 	if err != nil {
// 		log.Fatalf("SSH connection failed: %v", err)
// 	}

// 	client, err := goph.NewConn(&goph.Config{
// 		User:     s.User,
// 		Addr:     s.Host,
// 		Port:     uint(s.Port),
// 		Auth:     auth,
// 		Callback: ssh.InsecureIgnoreHostKey(),
// 	})
// 	if err != nil {
// 		log.Fatalf("SSH connection failed: %v", err)
// 	}

// 	defer client.Close()
// 	return client, nil
// }

func (s *SSH) RunCommand(cmd string) (string, error) {
	client, err := s.Connect()
	if err != nil {
		return "", err
	}
	defer client.Close()

	output, err := client.Run(cmd)
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

func (s *SSH) Terminal() {
	client, err := s.Connect()
	if err != nil {
		fmt.Print("Failed to connect to ssh")
		return
	}
	session, err := client.NewSession()
	if err != nil {
		fmt.Printf("Failed to create session: %s\n", err)
		return
	}
	defer session.Close()

	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	session.Stdin = os.Stdin

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	fileDescriptor := int(os.Stdin.Fd())
	if terminal.IsTerminal(fileDescriptor) {
		originalState, err := terminal.MakeRaw(fileDescriptor)
		if err != nil {
			panic(err)
		}
		defer terminal.Restore(fileDescriptor, originalState)

		termWidth, termHeight, err := terminal.GetSize(fileDescriptor)
		if err != nil {
			panic(err)
		}

		err = session.RequestPty("xterm-256color", termHeight, termWidth, modes)
		if err != nil {
			panic(err)
		}
	}

	err = session.Shell()
	if err != nil {
		return
	}
	session.Wait()
}
