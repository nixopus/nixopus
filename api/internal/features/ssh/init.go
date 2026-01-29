package ssh

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/melbahja/goph"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh/service"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

// SSH represents a single SSH connection configuration
type SSH struct {
	PrivateKey          string `json:"private_key"`
	PublicKey           string `json:"public_key"`
	Host                string `json:"host"`
	User                string `json:"user"`
	Port                uint   `json:"port"`
	Password            string `json:"password"`
	PrivateKeyProtected string `json:"private_key_protected"`
}

// connectionPoolEntry represents a pooled SSH connection with metadata
type connectionPoolEntry struct {
	client   *goph.Client
	lastUsed time.Time
	mu       sync.RWMutex
}

// SSHManager manages multiple SSH clients and provides a unified interface
// For now, it defaults to single client mode for backward compatibility
// In the future, it can be extended to support multiple clients/discoveries
type SSHManager struct {
	clients     map[string]*SSH                 // Map of client ID to SSH config
	defaultID   string                          // ID of the default client
	mu          sync.RWMutex                    // Mutex for thread safe access
	pool        map[string]*connectionPoolEntry // Connection pool by client ID
	poolMu      sync.RWMutex                    // Mutex for connection pool
	logger      logger.Logger                   // Logger for connection errors
	maxIdleTime time.Duration                   // Maximum idle time before closing connection
}

var (
	// orgManagers caches SSHManager instances per organization ID
	orgManagers   = make(map[string]*SSHManager)
	orgManagersMu sync.RWMutex
)

// GetSSHManagerForOrganization returns an SSHManager for a specific organization.
// Caches managers per organization to avoid repeated database queries.
// The manager is initialized with the active SSH key from the database for that organization.
func GetSSHManagerForOrganization(ctx context.Context, orgID uuid.UUID) (*SSHManager, error) {
	if config.GlobalStore == nil {
		return nil, fmt.Errorf("global store not initialized, ensure config.Init() has been called")
	}

	orgIDStr := orgID.String()

	// Check cache first
	orgManagersMu.RLock()
	if manager, exists := orgManagers[orgIDStr]; exists {
		orgManagersMu.RUnlock()
		return manager, nil
	}
	orgManagersMu.RUnlock()

	// Create new manager with organization-specific SSH config
	sshService := service.NewSSHKeyService(config.GlobalStore, ctx, logger.NewLogger())
	sshConfig, err := sshService.GetSSHConfigForOrganization(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH config for organization %s: %w", orgIDStr, err)
	}

	sshClient := NewSSHFromConfig(sshConfig)
	manager := NewSSHManager()
	manager.clients["default"] = sshClient

	// Cache manager
	orgManagersMu.Lock()
	orgManagers[orgIDStr] = manager
	orgManagersMu.Unlock()

	return manager, nil
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

// NewSSHManager creates a new SSH manager with a single default client from config
// For most use cases, prefer GetSSHManagerFromContext() to get an organization-specific manager
func NewSSHManager() *SSHManager {
	defaultClient := NewSSH()
	manager := &SSHManager{
		clients:     make(map[string]*SSH),
		defaultID:   "default",
		pool:        make(map[string]*connectionPoolEntry),
		logger:      logger.NewLogger(),
		maxIdleTime: 5 * time.Minute,
	}
	manager.clients["default"] = defaultClient
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

// ConnectWithID connects to a specific SSH client by ID with connection pooling
func (m *SSHManager) ConnectWithID(id string) (*goph.Client, error) {
	if id == "" {
		id = m.defaultID
	}

	// Try to get existing connection from pool
	m.poolMu.RLock()
	entry, exists := m.pool[id]
	var client *goph.Client
	if exists && entry != nil {
		entry.mu.RLock()
		client = entry.client
		entry.mu.RUnlock()
	}
	m.poolMu.RUnlock()

	// Validate the pooled connection is still alive
	if client != nil {
		if m.isConnectionAlive(client) {
			// Connection is valid, update last used time and return
			m.poolMu.Lock()
			if entry, exists := m.pool[id]; exists {
				entry.mu.Lock()
				entry.lastUsed = time.Now()
				entry.mu.Unlock()
			}
			m.poolMu.Unlock()
			return client, nil
		}
		// Connection is dead, remove it from pool
		m.poolMu.Lock()
		if entry, exists := m.pool[id]; exists {
			entry.mu.Lock()
			if entry.client == client {
				entry.client = nil
			}
			entry.mu.Unlock()
			delete(m.pool, id)
		}
		m.poolMu.Unlock()
	}

	// No valid pooled connection available, create new one
	sshClient, err := m.GetClient(id)
	if err != nil {
		return nil, err
	}

	client, err = sshClient.ConnectWithRetry()
	if err != nil {
		return nil, err
	}

	// Store in pool
	m.poolMu.Lock()
	m.pool[id] = &connectionPoolEntry{
		client:   client,
		lastUsed: time.Now(),
	}
	m.poolMu.Unlock()

	return client, nil
}

// cleanupIdleConnections periodically closes idle connections
func (m *SSHManager) cleanupIdleConnections() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		m.poolMu.Lock()
		for id, entry := range m.pool {
			entry.mu.Lock()
			if entry.client != nil && now.Sub(entry.lastUsed) > m.maxIdleTime {
				entry.client.Close()
				entry.client = nil
				delete(m.pool, id)
			}
			entry.mu.Unlock()
		}
		m.poolMu.Unlock()
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
		if entry.client != nil {
			entry.client.Close()
			entry.client = nil
		}
		entry.mu.Unlock()
		delete(m.pool, id)
	}
	m.poolMu.Unlock()
}

// RunCommand runs a command on the default SSH client
func (m *SSHManager) RunCommand(cmd string) (string, error) {
	return m.RunCommandWithID("", cmd)
}

// RunCommandWithID runs a command on a specific SSH client by ID
func (m *SSHManager) RunCommandWithID(id string, cmd string) (string, error) {
	client, err := m.GetClient(id)
	if err != nil {
		return "", err
	}
	return client.RunCommand(cmd)
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

// NewSSH creates a new SSH client from the global config
// This is the default way to create a single SSH client
func NewSSH() *SSH {
	return &SSH{
		PrivateKey:          config.AppConfig.SSH.PrivateKey,
		Host:                config.AppConfig.SSH.Host,
		User:                config.AppConfig.SSH.User,
		Port:                config.AppConfig.SSH.Port,
		Password:            config.AppConfig.SSH.Password,
		PrivateKeyProtected: config.AppConfig.SSH.PrivateKeyProtected,
	}
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
		return NewSSH() // Fallback to default config
	}
	return &SSH{
		PrivateKey:          sshConfig.PrivateKey,
		Host:                sshConfig.Host,
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

	// Try private key first
	client, err := s.ConnectWithPrivateKey()
	if err == nil {
		return client, nil
	}

	// If private key fails, try password with exponential backoff
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := time.Duration(attempt) * baseDelay
			time.Sleep(delay)
		}

		client, err = s.ConnectWithPassword()
		if err == nil {
			return client, nil
		}
	}

	return nil, fmt.Errorf("failed to connect with both private key and password after %d attempts: %w", maxRetries, err)
}

func (s *SSH) ConnectWithPrivateKey() (*goph.Client, error) {
	if s.PrivateKey == "" {
		return nil, fmt.Errorf("private key is required for SSH connection")
	}

	auth, err := goph.Key(s.PrivateKey, "")
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH auth from private key: %w", err)
	}

	client, err := goph.NewConn(&goph.Config{
		User:     s.User,
		Addr:     s.Host,
		Port:     uint(s.Port),
		Auth:     auth,
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
