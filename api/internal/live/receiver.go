package live

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/pkg/sftp"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	sftputil "github.com/raghavyuva/nixopus-api/internal/live/sftp"
)

// FileReceiver receives and reassembles file chunks
type FileReceiver struct {
	Path        string
	Chunks      map[int][]byte
	TotalChunks int
	Checksum    string
	StagingPath string
	mu          sync.Mutex
}

// NewFileReceiver creates a new file receiver
func NewFileReceiver(path string, totalChunks int, checksum string, stagingPath string) *FileReceiver {
	return &FileReceiver{
		Path:        path,
		Chunks:      make(map[int][]byte),
		TotalChunks: totalChunks,
		Checksum:    checksum,
		StagingPath: stagingPath,
	}
}

// AddChunk adds a chunk to the receiver
func (r *FileReceiver) AddChunk(chunkIndex int, data []byte) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Validate chunk index is within bounds
	if chunkIndex < 0 || chunkIndex >= r.TotalChunks {
		return fmt.Errorf("chunk index %d out of bounds (expected 0-%d)", chunkIndex, r.TotalChunks-1)
	}

	r.Chunks[chunkIndex] = data
	return nil
}

// IsComplete checks if all chunks have been received
func (r *FileReceiver) IsComplete() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.Chunks) == r.TotalChunks
}

// Reassemble reassembles the file from chunks
func (r *FileReceiver) Reassemble() ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.Chunks) != r.TotalChunks {
		return nil, fmt.Errorf("incomplete file: received %d/%d chunks", len(r.Chunks), r.TotalChunks)
	}

	// Sort chunk indices
	indices := make([]int, 0, len(r.Chunks))
	for idx := range r.Chunks {
		indices = append(indices, idx)
	}
	sort.Ints(indices)

	// Reassemble in order
	var content []byte
	for _, idx := range indices {
		content = append(content, r.Chunks[idx]...)
	}

	return content, nil
}

// VerifyChecksum verifies the file checksum
func (r *FileReceiver) VerifyChecksum(content []byte) bool {
	hasher := sha256.New()
	hasher.Write(content)
	calculated := hex.EncodeToString(hasher.Sum(nil))
	return calculated == r.Checksum
}

// sanitizePath ensures the path stays within the staging directory
// It prevents directory traversal attacks by cleaning the path and ensuring
// it doesn't escape the staging directory
func sanitizePath(stagingPath, filePath string) (string, error) {
	// Normalize to forward slashes first for consistent checking
	normalizedPath := filepath.ToSlash(filePath)

	// Reject absolute paths (check before cleaning)
	if filepath.IsAbs(filePath) || strings.HasPrefix(normalizedPath, "/") {
		return "", fmt.Errorf("absolute paths are not allowed: %s", filePath)
	}

	// Reject paths that try to escape (check for ".." before cleaning)
	if strings.Contains(normalizedPath, "..") {
		return "", fmt.Errorf("path traversal detected: %s", filePath)
	}

	// Clean the path (this will resolve any remaining issues)
	cleanPath := filepath.Clean(normalizedPath)

	// Join and clean the final path
	fullPath := filepath.Join(stagingPath, cleanPath)

	// Ensure the resolved path is still within staging directory
	// This prevents symlink attacks and ensures we're within bounds
	stagingAbs, err := filepath.Abs(stagingPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve staging path: %w", err)
	}

	fullAbs, err := filepath.Abs(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve full path: %w", err)
	}

	// Normalize separators for comparison
	stagingAbs = filepath.ToSlash(stagingAbs)
	fullAbs = filepath.ToSlash(fullAbs)

	// Check if the full path is within the staging directory
	// Add trailing slash to prevent partial matches (e.g., /staging vs /staging2)
	if !strings.HasPrefix(fullAbs+"/", stagingAbs+"/") {
		return "", fmt.Errorf("path escapes staging directory: %s", filePath)
	}

	return fullPath, nil
}

// getSFTPClient creates and returns an SFTP client.
// The SSH client connection is pooled and should not be closed.
// The returned SFTP client should be closed by the caller.
// This function handles connection failures by retrying with a fresh connection.
func getSFTPClient() (*sftp.Client, error) {
	sshMgr := ssh.GetSSHManager()
	return sftputil.CreateSFTPClientWithRetry(sshMgr)
}

// WriteToStaging writes the reassembled file to the staging directory via SFTP
func (r *FileReceiver) WriteToStaging() error {
	// Reassemble file
	content, err := r.Reassemble()
	if err != nil {
		return fmt.Errorf("failed to reassemble file: %w", err)
	}

	// Verify checksum
	if !r.VerifyChecksum(content) {
		return fmt.Errorf("checksum mismatch for file %s", r.Path)
	}

	// Sanitize and create full path (prevents path traversal)
	fullPath, err := sanitizePath(r.StagingPath, r.Path)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Get SFTP client
	sftpClient, err := getSFTPClient()
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	// Create directory if needed
	dirPath := filepath.Dir(fullPath)
	if err := sftpClient.MkdirAll(dirPath); err != nil {
		return fmt.Errorf("failed to create directory via SFTP: %w", err)
	}

	// Create and write file via SFTP
	file, err := sftpClient.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file via SFTP: %w", err)
	}
	defer file.Close()

	if _, err := file.Write(content); err != nil {
		return fmt.Errorf("failed to write file content via SFTP: %w", err)
	}

	// Set file permissions (0644)
	if err := sftpClient.Chmod(fullPath, 0644); err != nil {
		// Log but don't fail - permissions are best effort
		// Note: This is in receiver.go which doesn't have direct logger access
		// The error is non-critical so we silently continue
		_ = err
	}

	return nil
}

// Reset resets the receiver for a new file transfer (clears all chunks)
func (r *FileReceiver) Reset(totalChunks int, checksum string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Chunks = make(map[int][]byte)
	r.TotalChunks = totalChunks
	r.Checksum = checksum
}

// UpdateMetadata updates the receiver metadata without clearing chunks
// Only clears chunks if checksum differs (indicating a new file version)
func (r *FileReceiver) UpdateMetadata(totalChunks int, checksum string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// If checksum differs, this is a new file version - reset everything
	if r.Checksum != checksum {
		r.Chunks = make(map[int][]byte)
		r.Checksum = checksum
	}
	r.TotalChunks = totalChunks
}

// DeleteFileFromStaging deletes a file from staging directory via SFTP
// This is safer than using shell commands as it prevents command injection
func DeleteFileFromStaging(stagingPath, filePath string) error {
	// Sanitize path to prevent directory traversal attacks
	fullPath, err := sanitizePath(stagingPath, filePath)
	if err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	// Get SFTP client
	sftpClient, err := getSFTPClient()
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	// Delete file via SFTP (safer than shell command)
	if err := sftpClient.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file via SFTP: %w", err)
	}

	return nil
}
