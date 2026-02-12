package sftp

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/pkg/sftp"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
)

const (
	// maxRetries controls how many times SFTP client creation is retried on stale connections.
	maxRetries = 3
)

// isClosedConnectionError checks if the error indicates a closed or stale network connection.
// This includes EOF errors which occur when the remote SSH connection has been dropped.
func isClosedConnectionError(err error) bool {
	if err == nil {
		return false
	}
	if err == io.EOF {
		return true
	}
	errMsg := err.Error()
	return strings.Contains(errMsg, "use of closed network connection") ||
		strings.Contains(errMsg, "connection closed") ||
		strings.Contains(errMsg, "EOF") ||
		strings.Contains(errMsg, "broken pipe") ||
		strings.Contains(errMsg, "connection reset by peer")
}

// CreateSFTPClientWithRetry creates an SFTP client with automatic retry logic.
// It handles stale connections by removing them from the pool and retrying.
// The SSH client connection is pooled and should not be closed by the caller.
// The returned SFTP client should be closed by the caller.
func CreateSFTPClientWithRetry(sshMgr *ssh.SSHManager) (*sftp.Client, error) {
	for attempt := 0; attempt < maxRetries; attempt++ {
		client, err := sshMgr.Connect()
		if err != nil {
			return nil, fmt.Errorf("failed to connect via SSH: %w", err)
		}
		// Note: We don't close the client here as it's pooled and will be reused

		sftpClient, err := client.NewSftp()
		if err != nil {
			if isClosedConnectionError(err) {
				// Remove the bad connection from pool and retry
				sshMgr.CloseConnection("")
				if attempt < maxRetries-1 {
					continue
				}
			}
			return nil, fmt.Errorf("failed to create SFTP client: %w", err)
		}

		return sftpClient, nil
	}

	return nil, fmt.Errorf("failed to create SFTP client after %d attempts", maxRetries)
}

// ReadFile reads a file from the remote server via SFTP.
// This is a generic utility function that can be used across packages to read files remotely.
func ReadFile(ctx context.Context, filePath string) (string, error) {
	sshMgr, err := ssh.GetSSHManagerFromContext(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get SSH manager: %w", err)
	}

	sftpClient, err := CreateSFTPClientWithRetry(sshMgr)
	if err != nil {
		return "", err
	}
	defer sftpClient.Close()

	file, err := sftpClient.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file via SFTP: %w", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file content via SFTP: %w", err)
	}

	return string(content), nil
}

// FileExists checks if a file exists at the given path via SFTP.
// This is a generic utility function that can be used across packages to check file existence remotely.
// Note: For batch operations, use FilesExist instead.
func FileExists(ctx context.Context, path string) bool {
	sshMgr, err := ssh.GetSSHManagerFromContext(ctx)
	if err != nil {
		return false
	}

	sftpClient, err := CreateSFTPClientWithRetry(sshMgr)
	if err != nil {
		return false
	}
	defer sftpClient.Close()

	_, err = sftpClient.Stat(path)
	return err == nil
}

// markAllAsNonExistent is a helper function that marks all paths as non-existent.
func markAllAsNonExistent(paths []string) map[string]bool {
	result := make(map[string]bool, len(paths))
	for _, path := range paths {
		result[path] = false
	}
	return result
}

// FilesExist checks if multiple files exist in a single SFTP session.
// This is more efficient than calling FileExists multiple times.
// Returns a map of file path to existence boolean.
func FilesExist(ctx context.Context, paths []string) map[string]bool {
	sshMgr, err := ssh.GetSSHManagerFromContext(ctx)
	if err != nil {
		return markAllAsNonExistent(paths)
	}

	sftpClient, err := CreateSFTPClientWithRetry(sshMgr)
	if err != nil {
		return markAllAsNonExistent(paths)
	}
	defer sftpClient.Close()

	result := make(map[string]bool, len(paths))
	for _, path := range paths {
		_, err := sftpClient.Stat(path)
		result[path] = err == nil
	}

	return result
}
