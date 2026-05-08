package controller

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/features/agent/service"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/types"
)

type AgentController struct {
	service *service.AgentService
	logger  logger.Logger
}

func NewAgentController(store *storage.Store, ctx context.Context, l logger.Logger, notifier ...types.Notifier) *AgentController {
	var n types.Notifier
	if len(notifier) > 0 {
		n = notifier[0]
	}
	svc := service.NewAgentService(store, ctx, l, n)
	return &AgentController{
		service: svc,
		logger:  l,
	}
}
