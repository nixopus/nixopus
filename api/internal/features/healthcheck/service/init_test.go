package service

import (
	"context"
	"testing"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/stretchr/testify/assert"
)

func TestNewHealthCheckService(t *testing.T) {
	svc := NewHealthCheckService(nil, context.Background(), logger.NewLogger(), &mockHealthCheckRepo{})
	assert.NotNil(t, svc)
}
