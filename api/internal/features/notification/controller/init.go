package controller

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/service"
	"github.com/nixopus/nixopus/api/internal/features/notification/storage"
	"github.com/nixopus/nixopus/api/internal/features/notification/validation"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
)

type NotificationController struct {
	validator  *validation.Validator
	service    *service.NotificationService
	ctx        context.Context
	logger     logger.Logger
	dispatcher *notification.Dispatcher
}

func NewNotificationController(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
	dispatcher *notification.Dispatcher,
) *NotificationController {
	s := storage.NotificationStorage{DB: store.DB, Ctx: ctx}
	return &NotificationController{
		validator:  validation.NewValidator(&s),
		service:    service.NewNotificationService(store, ctx, l, &s),
		ctx:        ctx,
		logger:     l,
		dispatcher: dispatcher,
	}
}

func (c *NotificationController) parseAndValidate(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to read request body", err.Error())
		return false
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.validator.ParseRequestBody(r, io.NopCloser(bytes.NewBuffer(bodyBytes)), req); err != nil {
		c.logger.Log(logger.Error, "Failed to decode request", err.Error())
		return false
	}

	if err := c.validator.ValidateRequest(req); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		return false
	}

	return true
}
