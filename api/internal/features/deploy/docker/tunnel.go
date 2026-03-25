package docker

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/melbahja/goph"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/ssh"
)

const remoteDockerSocket = "/var/run/docker.sock"

// SSHTunnel forwards local Unix socket connections to a remote Docker daemon
// socket over a single, persistent SSH connection. Multiple Docker API calls
// multiplex as channels on the one TCP connection, avoiding SSH rate-limits.
type SSHTunnel struct {
	localSocket string
	sshClient   *ssh.SSH
	listener    net.Listener
	cleanup     func() error

	conn          *goph.Client
	connMu        sync.Mutex
	stopKeepalive chan struct{} // closed to stop the keepalive goroutine
}

// CreateSSHTunnel creates a local Unix socket and forwards all connections
// through the provided SSH client to the remote Docker daemon socket.
func CreateSSHTunnel(sshClient *ssh.SSH, lgr logger.Logger) (*SSHTunnel, error) {
	tempDir := os.TempDir()
	localSocket := filepath.Join(tempDir, fmt.Sprintf("docker-ssh-%d.sock", time.Now().UnixNano()))

	os.Remove(localSocket)

	listener, err := net.Listen("unix", localSocket)
	if err != nil {
		return nil, fmt.Errorf("failed to create local socket: %w", err)
	}

	tunnel := &SSHTunnel{
		localSocket: localSocket,
		sshClient:   sshClient,
		listener:    listener,
		cleanup: func() error {
			listener.Close()
			os.Remove(localSocket)
			return nil
		},
	}

	go tunnel.handleConnections(lgr)

	return tunnel, nil
}

// getConn returns the persistent SSH connection, creating it on first call.
// Starts an SSH keepalive goroutine to prevent NAT/firewall idle timeouts.
func (t *SSHTunnel) getConn() (*goph.Client, error) {
	t.connMu.Lock()
	defer t.connMu.Unlock()
	if t.conn != nil {
		return t.conn, nil
	}
	conn, err := t.sshClient.Connect()
	if err != nil {
		return nil, err
	}
	t.conn = conn
	t.stopKeepalive = make(chan struct{})
	ssh.StartKeepalive(conn, ssh.KeepaliveInterval, ssh.KeepaliveMaxMissed, t.stopKeepalive)
	return conn, nil
}

// resetConn closes the current connection and its keepalive so the next
// getConn creates a fresh one.
func (t *SSHTunnel) resetConn() {
	t.connMu.Lock()
	defer t.connMu.Unlock()
	if t.stopKeepalive != nil {
		close(t.stopKeepalive)
		t.stopKeepalive = nil
	}
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}
}

// isConnectionLevelError returns true for errors that indicate the SSH TCP
// connection itself is dead (EOF, broken pipe, reset). Returns false for
// channel-level errors like "open failed" which mean the remote resource
// (e.g. Docker socket) is unavailable but the SSH connection is fine.
func isConnectionLevelError(err error) bool {
	if err == nil {
		return false
	}
	if err == io.EOF {
		return true
	}
	msg := err.Error()
	return strings.Contains(msg, "EOF") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "use of closed network connection")
}

func (t *SSHTunnel) handleConnections(lgr logger.Logger) {
	for {
		localConn, err := t.listener.Accept()
		if err != nil {
			// Expected when Close() closes the listener (cache invalidation, replacement)
			if errors.Is(err, net.ErrClosed) || isConnectionLevelError(err) {
				return
			}
			lgr.Log(logger.Error, "SSH tunnel listener error", err.Error())
			return
		}

		go t.forwardConnection(localConn, lgr)
	}
}

// forwardConnection opens a channel on the persistent SSH connection to reach
// the remote Docker socket. If the channel open fails due to a stale TCP
// connection (EOF, broken pipe), it reconnects once and retries. Channel-level
// failures like "open failed" (Docker not running) do NOT trigger a reconnect.
func (t *SSHTunnel) forwardConnection(localConn net.Conn, lgr logger.Logger) {
	defer localConn.Close()

	conn, err := t.getConn()
	if err != nil {
		lgr.Log(logger.Error, "Failed to establish SSH connection", err.Error())
		return
	}

	remoteConn, err := conn.Dial("unix", remoteDockerSocket)
	if err != nil {
		if !isConnectionLevelError(err) {
			lgr.Log(logger.Error, "Failed to connect to remote Docker socket", err.Error())
			return
		}
		// Connection-level error: reset and retry once.
		t.resetConn()
		conn, err = t.getConn()
		if err != nil {
			lgr.Log(logger.Error, "Failed to re-establish SSH connection", err.Error())
			return
		}
		remoteConn, err = conn.Dial("unix", remoteDockerSocket)
		if err != nil {
			lgr.Log(logger.Error, "Failed to connect to remote Docker socket", err.Error())
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

// Close tears down the persistent SSH connection and the local socket.
func (t *SSHTunnel) Close() error {
	t.resetConn()
	if t.cleanup != nil {
		return t.cleanup()
	}
	return nil
}
