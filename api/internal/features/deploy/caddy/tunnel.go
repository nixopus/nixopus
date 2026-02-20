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
	"github.com/raghavyuva/caddygo"
	"github.com/raghavyuva/nixopus-api/internal/config"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// CaddyTunnel represents an SSH tunnel that forwards TCP connections
// to the remote Caddy admin API through SSH.
type CaddyTunnel struct {
	listener  net.Listener
	sshClient *ssh.SSH
	endpoint  string
	cleanup   func() error
}

// CreateCaddyTunnel creates a local TCP listener and forwards all connections
// through SSH to the remote Caddy admin API. The port is parsed from CADDY_ENDPOINT.
// Returns a CaddyTunnel with an endpoint suitable for caddygo.Client.
func CreateCaddyTunnel(sshClient *ssh.SSH, remotePort string, lgr logger.Logger) (*CaddyTunnel, error) {
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

func (t *CaddyTunnel) forwardConnection(localConn net.Conn, remotePort string, lgr logger.Logger) {
	defer localConn.Close()

	sshConn, err := t.sshClient.Connect()
	if err != nil {
		lgr.Log(logger.Error, "Caddy tunnel: failed to establish SSH connection", err.Error())
		return
	}
	defer sshConn.Close()

	remoteConn, err := sshConn.Dial("tcp", "127.0.0.1:"+remotePort)
	if err != nil {
		lgr.Log(logger.Error, "Caddy tunnel: failed to connect to remote Caddy", err.Error())
		return
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

// Close cleans up the tunnel by closing the listener.
func (t *CaddyTunnel) Close() error {
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

// GetCaddyClient returns a caddygo client that uses an SSH tunnel to reach
// the Caddy admin API on the given host. Uses existing SSH config (ctx org or sshClient).
// Port is parsed from CADDY_ENDPOINT. Caches tunnel per host+port for reuse.
func GetCaddyClient(ctx context.Context, sshClient *ssh.SSH, lgr *logger.Logger) (*caddygo.Client, error) {
	remotePort, err := parseCaddyEndpointPort()
	if err != nil {
		return nil, err
	}

	var s *ssh.SSH
	if sshClient != nil {
		s = sshClient
	} else {
		orgIDAny := ctx.Value(shared_types.OrganizationIDKey)
		if orgIDAny == nil {
			return nil, fmt.Errorf("organization ID or SSH client required for Caddy")
		}
		var orgID uuid.UUID
		switch v := orgIDAny.(type) {
		case string:
			var err error
			orgID, err = uuid.Parse(v)
			if err != nil {
				return nil, fmt.Errorf("invalid organization ID: %w", err)
			}
		case uuid.UUID:
			orgID = v
		default:
			return nil, fmt.Errorf("unexpected organization ID type: %T", orgIDAny)
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
	tunnel, err := CreateCaddyTunnel(s, remotePort, *lgr)
	if err != nil {
		return nil, err
	}

	client := caddygo.NewClient(tunnel.Endpoint())

	caddyTunnelCacheMu.Lock()
	caddyTunnelCache[key] = &caddyTunnelEntry{tunnel: tunnel, client: client, lastUsed: time.Now()}
	caddyTunnelCacheMu.Unlock()

	return client, nil
}
