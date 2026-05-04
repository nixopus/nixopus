package testutil

import "github.com/nixopus/nixopus/api/internal/features/logger"

// NewLogger returns a standard logger for use in tests.
func NewLogger() logger.Logger { return logger.NewLogger() }
