package caddy

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/melbahja/goph"
	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// CaddyTunnel represents an SSH tunnel that forwards TCP connections
// to the remote Caddy admin API through SSH. It holds a persistent SSH
// connection and multiplexes all forwarded traffic over it.
type CaddyTunnel struct {
	listener      net.Listener
	sshClient     *ssh.SSH
	endpoint      string
	cleanup       func() error
	connMu        sync.Mutex
	persistSSH    *goph.Client
	stopKeepalive chan struct{} // closed to stop the keepalive goroutine
	orgID         uuid.UUID     // for debug logging when tunnel fails
}

// CreateCaddyTunnel creates a local TCP listener and forwards all connections
// through SSH to the remote Caddy admin API. The port is parsed from CADDY_ENDPOINT.
// Returns a CaddyTunnel with an endpoint suitable for caddygo.Client.
// orgID is used for debug logging when the tunnel fails (pass uuid.Nil if unknown).
func CreateCaddyTunnel(sshClient *ssh.SSH, remotePort string, orgID uuid.UUID, lgr logger.Logger) (*CaddyTunnel, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create caddy tunnel listener: %w", err)
	}

	addr := listener.Addr().String()
	endpoint := "http://" + addr

	tunnel := &CaddyTunnel{
		listener:  listener,
		sshClient: sshClient,
		endpoint:  endpoint,
		orgID:     orgID,
		cleanup: func() error {
			return listener.Close()
		},
	}

	go tunnel.handleConnections(remotePort, lgr)

	return tunnel, nil
}

func (t *CaddyTunnel) handleConnections(remotePort string, lgr logger.Logger) {
	for {
		localConn, err := t.listener.Accept()
		if err != nil {
			lgr.Log(logger.Error, "Caddy tunnel listener error", err.Error())
			return
		}

		go t.forwardConnection(localConn, remotePort, lgr)
	}
}

// getOrCreateSSH returns a persistent SSH connection, reconnecting on failure.
// All forwarded connections multiplex over this single SSH TCP connection.
// Starts an SSH keepalive goroutine to prevent NAT/firewall idle timeouts.
// Stale connections are detected when Dial fails.
func (t *CaddyTunnel) getOrCreateSSH() (*goph.Client, error) {
	t.connMu.Lock()
	defer t.connMu.Unlock()

	if t.persistSSH != nil {
		return t.persistSSH, nil
	}

	client, err := t.sshClient.Connect()
	if err != nil {
		return nil, err
	}
	t.persistSSH = client
	t.stopKeepalive = make(chan struct{})
	ssh.StartKeepalive(client, ssh.KeepaliveInterval, ssh.KeepaliveMaxMissed, t.stopKeepalive)
	return client, nil
}

func (t *CaddyTunnel) tunnelErrLog(hostLabel, errMsg string) string {
	if t.orgID != uuid.Nil {
		return fmt.Sprintf("host=%s org=%s err=%s", hostLabel, t.orgID.String(), errMsg)
	}
	return fmt.Sprintf("host=%s err=%s", hostLabel, errMsg)
}

func (t *CaddyTunnel) forwardConnection(localConn net.Conn, remotePort string, lgr logger.Logger) {
	defer localConn.Close()

	hostLabel := t.sshClient.Host
	if t.sshClient.Port != 0 && t.sshClient.Port != 22 {
		hostLabel = fmt.Sprintf("%s:%d", t.sshClient.Host, t.sshClient.Port)
	}

	sshConn, err := t.getOrCreateSSH()
	if err != nil {
		lgr.Log(logger.Error, "Caddy tunnel: failed to establish SSH connection", t.tunnelErrLog(hostLabel, err.Error()))
		return
	}

	remoteConn, err := sshConn.Dial("tcp", "127.0.0.1:"+remotePort)
	if err != nil {
		// Connection may have gone stale between getOrCreateSSH and Dial.
		// Stop keepalive, invalidate, and retry once.
		t.connMu.Lock()
		if t.persistSSH == sshConn {
			if t.stopKeepalive != nil {
				close(t.stopKeepalive)
				t.stopKeepalive = nil
			}
			t.persistSSH.Close()
			t.persistSSH = nil
		}
		t.connMu.Unlock()

		sshConn, err = t.getOrCreateSSH()
		if err != nil {
			lgr.Log(logger.Error, "Caddy tunnel: failed to reconnect SSH", t.tunnelErrLog(hostLabel, err.Error()))
			return
		}
		remoteConn, err = sshConn.Dial("tcp", "127.0.0.1:"+remotePort)
		if err != nil {
			lgr.Log(logger.Error, "Caddy tunnel: failed to connect to remote Caddy after reconnect", t.tunnelErrLog(hostLabel, err.Error()))
			return
		}
	}
	defer remoteConn.Close()

	done := make(chan struct{}, 2)

	go func() {
		io.Copy(remoteConn, localConn)
		done <- struct{}{}
	}()

	go func() {
		io.Copy(localConn, remoteConn)
		done <- struct{}{}
	}()

	<-done
}

// Endpoint returns the local HTTP endpoint for the Caddy admin API.
func (t *CaddyTunnel) Endpoint() string {
	return t.endpoint
}

// Close cleans up the tunnel by stopping keepalive, closing the persistent
// SSH connection, and closing the listener.
func (t *CaddyTunnel) Close() error {
	t.connMu.Lock()
	if t.stopKeepalive != nil {
		close(t.stopKeepalive)
		t.stopKeepalive = nil
	}
	if t.persistSSH != nil {
		t.persistSSH.Close()
		t.persistSSH = nil
	}
	t.connMu.Unlock()

	if t.cleanup != nil {
		return t.cleanup()
	}
	return nil
}

// caddyTunnelCache caches Caddy tunnels and clients per SSH host with TTL-based
// eviction. Idle tunnels are cleaned up to avoid holding thousands of SSH
// connections open at scale.
var (
	caddyTunnelCache   = make(map[string]*caddyTunnelEntry)
	caddyTunnelCacheMu sync.RWMutex
	tunnelCleanupOnce  sync.Once
)

const tunnelIdleTTL = 10 * time.Minute

type caddyTunnelEntry struct {
	tunnel   *CaddyTunnel
	client   *caddygo.Client
	lastUsed time.Time
}

func startTunnelCleanup() {
	tunnelCleanupOnce.Do(func() {
		go func() {
			ticker := time.NewTicker(1 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				evictIdleTunnels()
			}
		}()
	})
}

func evictIdleTunnels() {
	caddyTunnelCacheMu.Lock()
	defer caddyTunnelCacheMu.Unlock()

	now := time.Now()
	for key, entry := range caddyTunnelCache {
		if now.Sub(entry.lastUsed) > tunnelIdleTTL {
			if entry.tunnel != nil {
				entry.tunnel.Close()
			}
			delete(caddyTunnelCache, key)
		}
	}
}

const defaultCaddyPort = "2019"

// parseCaddyEndpointPort extracts the port from CADDY_ENDPOINT (e.g. http://host:2019 -> "2019").
// Falls back to 2019 when not set or when port is omitted.
func parseCaddyEndpointPort() (string, error) {
	endpoint := config.AppConfig.Proxy.CaddyEndpoint
	if endpoint == "" {
		return defaultCaddyPort, nil
	}
	u, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid CADDY_ENDPOINT: %w", err)
	}
	port := u.Port()
	if port == "" {
		return defaultCaddyPort, nil
	}
	if _, err := strconv.Atoi(port); err != nil {
		return "", fmt.Errorf("invalid port in CADDY_ENDPOINT: %s", port)
	}
	return port, nil
}

// orgIDFromContext extracts organization ID from context for debug logging.
// Returns uuid.Nil if not present or invalid.
func orgIDFromContext(ctx context.Context) uuid.UUID {
	orgIDAny := ctx.Value(shared_types.OrganizationIDKey)
	if orgIDAny == nil {
		return uuid.Nil
	}
	switch v := orgIDAny.(type) {
	case string:
		id, err := uuid.Parse(v)
		if err != nil {
			return uuid.Nil
		}
		return id
	case uuid.UUID:
		return v
	default:
		return uuid.Nil
	}
}

// GetCaddyClient returns a caddygo client that uses an SSH tunnel to reach
// the Caddy admin API on the given host. Uses existing SSH config (ctx org or sshClient).
// Port is parsed from CADDY_ENDPOINT. Caches tunnel per host+port for reuse.
func GetCaddyClient(ctx context.Context, sshClient *ssh.SSH, lgr *logger.Logger) (*caddygo.Client, error) {
	remotePort, err := parseCaddyEndpointPort()
	if err != nil {
		return nil, err
	}

	var s *ssh.SSH
	var orgID uuid.UUID
	if sshClient != nil {
		s = sshClient
		orgID = orgIDFromContext(ctx)
	} else {
		orgID = orgIDFromContext(ctx)
		if orgID == uuid.Nil {
			return nil, fmt.Errorf("organization ID or SSH client required for Caddy")
		}
		manager, err := ssh.GetSSHManagerForOrganization(ctx, orgID)
		if err != nil {
			return nil, fmt.Errorf("failed to get SSH manager: %w", err)
		}
		s, err = manager.GetDefaultSSH()
		if err != nil {
			return nil, fmt.Errorf("failed to get SSH client: %w", err)
		}
	}

	key := s.Host + ":" + remotePort
	if s.Host == "" {
		return nil, fmt.Errorf("SSH host is required for Caddy tunnel")
	}

	startTunnelCleanup()

	caddyTunnelCacheMu.RLock()
	entry, exists := caddyTunnelCache[key]
	caddyTunnelCacheMu.RUnlock()

	if exists && entry != nil && entry.tunnel != nil {
		caddyTunnelCacheMu.Lock()
		entry.lastUsed = time.Now()
		caddyTunnelCacheMu.Unlock()
		return entry.client, nil
	}

	if lgr == nil {
		def := logger.NewLogger()
		lgr = &def
	}
	tunnel, err := CreateCaddyTunnel(s, remotePort, orgID, *lgr)
	if err != nil {
		return nil, err
	}

	client := caddygo.NewClient(tunnel.Endpoint())

	caddyTunnelCacheMu.Lock()
	caddyTunnelCache[key] = &caddyTunnelEntry{tunnel: tunnel, client: client, lastUsed: time.Now()}
	caddyTunnelCacheMu.Unlock()

	return client, nil
}
