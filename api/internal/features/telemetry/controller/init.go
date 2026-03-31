package controller

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/telemetry/service"
	"github.com/nixopus/nixopus/api/internal/features/telemetry/storage"
	"github.com/nixopus/nixopus/api/internal/features/telemetry/types"
	"github.com/nixopus/nixopus/api/internal/features/telemetry/validation"
	"github.com/uptrace/bun"
)

type TelemetryController struct {
	service   *service.TelemetryService
	validator *validation.Validator
	logger    logger.Logger
}

func NewTelemetryController(db *bun.DB, ctx context.Context, l logger.Logger) *TelemetryController {
	repo := storage.NewTelemetryStorage(db, ctx)
	return &TelemetryController{
		service:   service.NewTelemetryService(repo, ctx, l),
		validator: validation.NewValidator(),
		logger:    l,
	}
}

func (c *TelemetryController) HandleTrackInstall(f fuego.ContextWithBody[types.TrackInstallRequest]) (*types.TrackInstallResponse, error) {
	body, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, "failed to parse telemetry request body", err.Error())
		return nil, fuego.BadRequestError{Detail: "invalid request body", Err: err}
	}

	if err := c.validator.ValidateRequest(&body); err != nil {
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	clientIP := extractClientIP(f.Request())

	if err := c.service.TrackInstall(&body, clientIP); err != nil {
		c.logger.Log(logger.Error, "failed to track install event", err.Error())
		return nil, fuego.HTTPError{Err: err, Detail: "failed to record event", Status: http.StatusInternalServerError}
	}

	return &types.TrackInstallResponse{
		Status:  "success",
		Message: "event recorded",
	}, nil
}

func extractClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-Ip"); xri != "" {
		return strings.TrimSpace(xri)
	}
	ip := r.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return strings.Trim(ip, "[]")
}
