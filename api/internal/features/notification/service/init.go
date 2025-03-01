package service

import (
	"context"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification/storage"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
)

type NotificationService struct {
	storage storage.NotificationStorage
	Ctx     context.Context
	store   *shared_storage.Store
	logger  logger.Logger
}

func NewNotificationService(store *shared_storage.Store, ctx context.Context, logger logger.Logger) *NotificationService {
	return &NotificationService{
		storage: storage.NotificationStorage{
			DB:  store.DB,
			Ctx: ctx,
		},
		store:  store,
		Ctx:    ctx,
		logger: logger,
	}
}
