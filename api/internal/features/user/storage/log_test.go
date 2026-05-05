package storage

import (
	"testing"

	"github.com/nixopus/nixopus/api/internal/features/logger"
)

func TestStorageLog_NilLogger(t *testing.T) {
	storageLog(nil, logger.Info, "msg", "data")
}

func TestStorageLog_WithLogger(t *testing.T) {
	l := logger.NewLogger()
	storageLog(&l, logger.Info, "msg", "data")
}
