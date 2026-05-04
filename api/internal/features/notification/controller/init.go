package controller

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/notification"
	"github.com/nixopus/nixopus/api/internal/features/notification/service"
	"github.com/nixopus/nixopus/api/internal/features/notification/storage"
	"github.com/nixopus/nixopus/api/internal/features/notification/validation"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	"github.com/nixopus/nixopus/api/internal/utils"
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
	s := storage.NotificationStorage{DB: store.DB, Ctx: ctx, Logger: &l}
	return &NotificationController{
		validator:  validation.NewValidatorWithLogger(&s, &l),
		service:    service.NewNotificationService(store, ctx, l, &s),
		ctx:        ctx,
		logger:     l,
		dispatcher: dispatcher,
	}
}

func (c *NotificationController) parseAndValidate(w http.ResponseWriter, r *http.Request, req interface{}) bool {
	data := notificationRequestLogData(r, utils.GetUser(w, r))

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("notification: parseAndValidate: read body: %v", err), data)
		return false
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	if err := c.validator.ParseRequestBody(r, io.NopCloser(bytes.NewBuffer(bodyBytes)), req); err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("notification: parseAndValidate: decode JSON: %v", err), data)
		return false
	}

	if err := c.validator.ValidateRequest(req); err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("notification: parseAndValidate: validation: %v", err), data)
		return false
	}

	return true
}
