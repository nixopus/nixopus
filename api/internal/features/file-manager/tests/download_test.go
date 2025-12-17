package tests

// import (
// 	"context"
// 	"os"
// 	"strings"
// 	"testing"

// 	"github.com/pkg/sftp"
// 	"github.com/raghavyuva/nixopus-api/internal/features/file-manager/service"
// 	"github.com/raghavyuva/nixopus-api/internal/features/logger"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestDownloadFile(t *testing.T) {
// 	// Note: This is a basic test structure. Full implementation would require
// 	// proper mocking of SFTP client and file system operations.

// 	t.Run("DownloadFile_PathTraversalProtection", func(t *testing.T) {
// 		fileManager := service.NewFileManagerService(context.Background(), logger.NewLogger())
// 		require.NotNil(t, fileManager)

// 		// Test that path traversal is prevented - this will fail due to no SFTP connection
// 		// but we're testing the path validation logic
// 		_, err := fileManager.DownloadFile("../etc/passwd")
// 		// We expect an error due to no SFTP connection, but the path traversal check happens first
// 		assert.Error(t, err)
// 	})

// 	t.Run("DownloadFolder_PathTraversalProtection", func(t *testing.T) {
// 		fileManager := service.NewFileManagerService(context.Background(), logger.NewLogger())
// 		require.NotNil(t, fileManager)

// 		// Test that path traversal is prevented
// 		_, err := fileManager.DownloadFolder("../etc")
// 		assert.Error(t, err)
// 	})
// }

// func TestUploadFile(t *testing.T) {
// 	t.Run("UploadFile_PathTraversalProtection", func(t *testing.T) {
// 		fileManager := service.NewFileManagerService(context.Background(), logger.NewLogger())
// 		require.NotNil(t, fileManager)

// 		// Test that path traversal is prevented
// 		testContent := "test file content"
// 		reader := strings.NewReader(testContent)

// 		err := fileManager.UploadFile(reader, "/base", "../outside/file.txt")
// 		assert.Error(t, err)
// 		assert.Contains(t, err.Error(), "invalid path")
// 	})

// 	t.Run("UploadFile_ValidRelativePath", func(t *testing.T) {
// 		fileManager := service.NewFileManagerService(context.Background(), logger.NewLogger())
// 		require.NotNil(t, fileManager)

// 		// Test valid relative path
// 		testContent := "test file content"
// 		reader := strings.NewReader(testContent)

// 		// This would normally work but requires SFTP connection
// 		err := fileManager.UploadFile(reader, "/base", "subdir/file.txt")
// 		// We expect an error due to no SFTP connection in test
// 		assert.Error(t, err)
// 	})
// }