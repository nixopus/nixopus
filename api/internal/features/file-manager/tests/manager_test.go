package tests

// import (
// 	"context"
// 	"fmt"
// 	"os"
// 	"testing"

// 	"github.com/raghavyuva/nixopus-api/internal/features/file-manager/service"
// 	"github.com/raghavyuva/nixopus-api/internal/features/logger"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/require"
// )

// func TestManager(t *testing.T) {
// 	testLogger := func(format string, args ...interface{}) {
// 		logMsg := fmt.Sprintf(format, args...)
// 		fmt.Println(logMsg)
// 		t.Log(logMsg)
// 	}

// 	t.Run("ListFiles_RootPath", func(t *testing.T) {
// 		path := "/"
// 		testLogger("Testing ListFiles with path: %s", path)

// 		fileManager := service.NewFileManagerService(context.Background(),logger.NewLogger())
// 		require.NotNil(t, fileManager, "File manager instance should not be nil")

// 		files, err := fileManager.ListFiles(path)

// 		if err != nil {
// 			testLogger("Error occurred: %v", err)
// 		} else {
// 			testLogger("Found %d files in path %s", len(files), path)
// 			for i, file := range files {
// 				testLogger("File %d: %+v", i, file)
// 			}
// 		}

// 		assert.NoError(t, err, "ListFiles should not return an error for valid path")
// 		assert.NotNil(t, files, "Files list should not be nil")
// 	})

// 	t.Run("ListFiles_InvalidPath", func(t *testing.T) {
// 		path := "/non-existent-path"
// 		testLogger("Testing ListFiles with invalid path: %s", path)

// 		fileManager := service.NewFileManagerService(context.Background(),logger.NewLogger())
// 		files, err := fileManager.ListFiles(path)

// 		testLogger("Result: files=%v, err=%v", files, err)

// 		assert.Error(t, err, "ListFiles should return an error for invalid path")
// 	})

// 	t.Run("ListFiles_WithFilters", func(t *testing.T) {
// 		path := "/"
// 		testLogger("Testing ListFiles with filters on path: %s", path)

// 		fileManager := service.NewFileManagerService(context.Background(),logger.NewLogger())

// 		files, err := fileManager.ListFiles(path)

// 		testLogger("Result: found %d files, err=%v", len(files), err)

// 		assert.NoError(t, err)
// 	})
// }

// func TestFileManager_Integration(t *testing.T) {
// 	if testing.Short() {
// 		t.Skip("Skipping integration test in short mode")
// 	}

// 	integLogger := func(format string, args ...interface{}) {
// 		logMsg := fmt.Sprintf(format, args...)
// 		fmt.Println(logMsg)
// 		t.Log(logMsg)
// 	}

// 	tempDir, err := os.MkdirTemp("", "nixopus-test-*")
// 	if err != nil {
// 		t.Fatalf("Failed to create temp directory: %v", err)
// 	}
// 	defer os.RemoveAll(tempDir)

// 	integLogger("Created temporary test directory: %s", tempDir)

// 	testFiles := []string{"file1.txt", "file2.log", "file3.json"}
// 	for _, filename := range testFiles {
// 		filepath := fmt.Sprintf("%s/%s", tempDir, filename)
// 		if err := os.WriteFile(filepath, []byte("test content"), 0644); err != nil {
// 			t.Fatalf("Failed to create test file %s: %v", filename, err)
// 		}
// 		integLogger("Created test file: %s", filepath)
// 	}

// 	fileManager := service.NewFileManagerService(context.Background(), logger.NewLogger())
// 	files, err := fileManager.ListFiles(tempDir)

// 	integLogger("ListFiles result for temp directory: files=%d, err=%v", len(files), err)

// 	assert.NoError(t, err)
// 	assert.Equal(t, len(testFiles), len(files), "Number of files should match")

// 	for _, file := range files {
// 		integLogger("Verifying file: %+v", file)
// 	}
// }
