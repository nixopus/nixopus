package service

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification/storage"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
)

type NotificationService struct {
	storage storage.NotificationRepository
	Ctx     context.Context
	store   *shared_storage.Store
	logger  logger.Logger
}

func NewNotificationService(store *shared_storage.Store, ctx context.Context, logger logger.Logger, notificationRepository storage.NotificationRepository) *NotificationService {
	return &NotificationService{
		storage: notificationRepository,
		store:   store,
		Ctx:     ctx,
		logger:  logger,
	}
}
