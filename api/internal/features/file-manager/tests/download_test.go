package tests

import (
	"context"
import (
	"context"
	"strings"
	"testing"
	"github.com/pkg/sftp"
	"api/internal/features/file-manager/controller"
	"api/internal/features/file-manager/service"
	"api/internal/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)
	"testing"

	"github.com/pkg/sftp"
	"github.com/raghavyuva/nixopus-api/internal/features/file-manager/controller"
	"github.com/raghavyuva/nixopus-api/internal/features/file-manager/service"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDownloadFile(t *testing.T) {
	// Note: This is a basic test structure. Full implementation would require
	// proper mocking of SFTP client and file system operations.

	t.Run("DownloadFile_PathTraversalProtection", func(t *testing.T) {
		fileManager := service.NewFileManagerService(context.Background(), logger.NewLogger())
		require.NotNil(t, fileManager)

		// Test that path traversal is prevented - this will fail due to no SFTP connection
		// but we're testing the path validation logic
		_, err := fileManager.DownloadFile("../etc/passwd")
		// We expect an error due to no SFTP connection, but the path traversal check happens first
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid path")
	})

	t.Run("DownloadFolder_PathTraversalProtection", func(t *testing.T) {
		fileManager := service.NewFileManagerService(context.Background(), logger.NewLogger())
		require.NotNil(t, fileManager)

		// Test that path traversal is prevented
		_, err := fileManager.DownloadFolder("../etc")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid path")
	})
}

func TestUploadFile(t *testing.T) {
	t.Run("UploadFile_PathTraversalProtection", func(t *testing.T) {
		fileManager := service.NewFileManagerService(context.Background(), logger.NewLogger())
		require.NotNil(t, fileManager)

		// Test that path traversal is prevented
		testContent := "test file content"
		reader := strings.NewReader(testContent)

		err := fileManager.UploadFile(reader, "/base", "../outside/file.txt")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid path")
	})

	t.Run("UploadFile_ValidRelativePath", func(t *testing.T) {
		fileManager := service.NewFileManagerService(context.Background(), logger.NewLogger())
		require.NotNil(t, fileManager)

		// Test valid relative path
		testContent := "test file content"
		reader := strings.NewReader(testContent)

		// This would normally work but requires SFTP connection
		err := fileManager.UploadFile(reader, "/base", "subdir/file.txt")
		// We expect an error due to no SFTP connection in test
		assert.Error(t, err)
	})
}

func TestSanitizeFilenameForHeader(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "safe ASCII filename",
			input:    "document.pdf",
			expected: "attachment; filename=\"document.pdf\"",
		},
		{
			name:     "filename with spaces",
			input:    "my document.pdf",
			expected: "attachment; filename=\"my document.pdf\"",
		},
		{
			name:     "filename with dangerous characters",
			input:    "file\r\nwith\nbad\"chars\";evil",
			expected: "attachment; filename=\"filewithbadcharsevil\"",
		},
		{
			name:     "empty filename",
			input:    "",
			expected: "attachment; filename=\"download\"",
		},
		{
			name:     "filename with only dangerous chars",
			input:    "\"\r\n;;\\",
			expected: "attachment; filename=\"download\"",
		},
		{
			name:     "UTF-8 filename",
			input:    "файл.txt",
			expected: "attachment; filename=\"______.txt\"; filename*=UTF-8''%D1%84%D0%B0%D0%B9%D0%BB.txt",
		},
		{
			name:     "very long filename",
			input:    strings.Repeat("a", 300),
			expected: "attachment; filename=\"" + strings.Repeat("a", 255) + "\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := controller.SanitizeFilenameForHeader(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}