package controller

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/features/agent/service"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/storage"
)

type AgentController struct {
	service *service.AgentService
	logger  logger.Logger
}

func NewAgentController(store *storage.Store, ctx context.Context, l logger.Logger) *AgentController {
	svc := service.NewAgentService(store, ctx, l)
	return &AgentController{
		service: svc,
		logger:  l,
	}
}
