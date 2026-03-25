package utils

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/nixopus/nixopus/api/internal/features/ssh"
	"github.com/pkg/sftp"
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
		strings.Contains(errMsg, "connection lost") ||
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
// Uses org-scoped SFTP pool when context has OrganizationIDKey.
func ReadFile(ctx context.Context, filePath string) (string, error) {
	var content string
	err := WithSFTPClientFromPool(ctx, func(sftpClient *sftp.Client) error {
		file, err := sftpClient.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file via SFTP: %w", err)
		}
		defer file.Close()
		data, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read file content via SFTP: %w", err)
		}
		content = string(data)
		return nil
	})
	if err != nil {
		return "", err
	}
	return content, nil
}

// ReadFileBytes reads a file from the remote server via SFTP, returning raw bytes.
// Uses org-scoped SFTP pool when context has OrganizationIDKey.
func ReadFileBytes(ctx context.Context, filePath string) ([]byte, error) {
	var data []byte
	err := WithSFTPClientFromPool(ctx, func(sftpClient *sftp.Client) error {
		file, err := sftpClient.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file via SFTP: %w", err)
		}
		defer file.Close()
		data, err = io.ReadAll(file)
		return err
	})
	if err != nil {
		return nil, err
	}
	return data, nil
}

// WithSFTPClient runs fn with an SFTP client from the org-scoped pool.
// Context must have types.OrganizationIDKey set (required for GetSSHManagerFromContext).
// Reuses pooled connections per organization to reduce connection churn.
func WithSFTPClient(ctx context.Context, fn func(*sftp.Client) error) error {
	return WithSFTPClientFromPool(ctx, fn)
}

// ReadFileBytesFromClient reads a file via an existing SFTP client. Prefer this inside WithSFTPClient
// to avoid creating multiple connections.
func ReadFileBytesFromClient(client *sftp.Client, filePath string) ([]byte, error) {
	file, err := client.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file via SFTP: %w", err)
	}
	defer file.Close()
	return io.ReadAll(file)
}

// WalkRemote walks the directory tree at root via SFTP using a single connection.
// walkFn receives the client so it can read files without opening new connections.
// walkFn can return filepath.SkipDir to skip a directory.
// Context must have SSH manager set (e.g. OrganizationIDKey for tenant resolution).
func WalkRemote(ctx context.Context, root string, walkFn func(client *sftp.Client, path string, info os.FileInfo, err error) error) error {
	return WithSFTPClient(ctx, func(client *sftp.Client) error {
		walker := client.Walk(root)
		for walker.Step() {
			if err := walker.Err(); err != nil {
				return walkFn(client, walker.Path(), nil, err)
			}
			path := walker.Path()
			stat := walker.Stat()
			if err := walkFn(client, path, stat, nil); err != nil {
				if err == filepath.SkipDir && stat != nil && stat.IsDir() {
					walker.SkipDir()
					continue
				}
				return err
			}
		}
		return nil
	})
}

// FileExists checks if a file exists at the given path via SFTP.
// Uses org-scoped SFTP pool when context has OrganizationIDKey.
func FileExists(ctx context.Context, path string) bool {
	var exists bool
	_ = WithSFTPClientFromPool(ctx, func(sftpClient *sftp.Client) error {
		_, err := sftpClient.Stat(path)
		exists = err == nil
		return nil
	})
	return exists
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
// Uses org-scoped SFTP pool. Returns a map of file path to existence boolean.
func FilesExist(ctx context.Context, paths []string) map[string]bool {
	result := make(map[string]bool, len(paths))
	err := WithSFTPClientFromPool(ctx, func(sftpClient *sftp.Client) error {
		for _, path := range paths {
			_, err := sftpClient.Stat(path)
			result[path] = err == nil
		}
		return nil
	})
	if err != nil {
		return markAllAsNonExistent(paths)
	}
	return result
}
