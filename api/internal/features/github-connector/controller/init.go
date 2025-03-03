package controller

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/service"
	"github.com/raghavyuva/nixopus-api/internal/features/github-connector/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type GithubConnectorController struct {
	store        *shared_storage.Store
	validator    *validation.Validator
	service      *service.GithubConnectorService
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
}

func NewGithubConnectorController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	notificationManager *notification.NotificationManager,
) *GithubConnectorController {
	return &GithubConnectorController{
		store:        store,
		validator:    validation.NewValidator(),
		service:      service.NewGithubConnectorService(store, ctx, l),
		ctx:          ctx,
		logger:       l,
		notification: notificationManager,
	}
}
