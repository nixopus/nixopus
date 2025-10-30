package docker

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
)

// CreateSSHTunnel creates a local Unix socket and forwards all connections
// through the provided SSH client to the remote Docker daemon socket
// at /var/run/docker.sock. It returns an SSHTunnel with a cleanup function
// that closes the listener and removes the temporary socket file.
func CreateSSHTunnel(sshClient *ssh.SSH, logger logger.Logger) (*SSHTunnel, error) {
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

	go tunnel.handleConnections(logger)

	return tunnel, nil
}

// handleConnections manages the SSH tunnel connections
func (t *SSHTunnel) handleConnections(lgr logger.Logger) {
	for {
		localConn, err := t.listener.Accept()
		if err != nil {
			lgr.Log(logger.Error, "SSH tunnel listener error", err.Error())
			return
		}

		go t.forwardConnection(localConn, lgr)
	}
}

// forwardConnection forwards a local connection through the SSH tunnel to the remote Docker socket
func (t *SSHTunnel) forwardConnection(localConn net.Conn, lgr logger.Logger) {
	defer localConn.Close()

	sshConn, err := t.sshClient.Connect()
	if err != nil {
		lgr.Log(logger.Error, "Failed to establish SSH connection", err.Error())
		return
	}
	defer sshConn.Close()

	remoteConn, err := sshConn.Dial("unix", "/var/run/docker.sock")
	if err != nil {
		lgr.Log(logger.Error, "Failed to connect to remote Docker socket", err.Error())
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

// Close cleans up the DockerService and any SSH tunnels
func (s *DockerService) Close() error {
	if s.sshTunnel != nil {
		return s.sshTunnel.cleanup()
	}
	return nil
}
