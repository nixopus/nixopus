package auth

import (
	"context"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/service"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

type AuthController struct {
	validator    *validation.Validator
	service      service.AuthServiceInterface
	ctx          context.Context
	logger       logger.Logger
	notification *notification.NotificationManager
}

func NewAuthController(
	ctx context.Context,
	logger logger.Logger,
	notificationManager *notification.NotificationManager,
	authService service.AuthService,
) *AuthController {
	return &AuthController{
		validator:    validation.NewValidator(),
		service:      &authService,
		ctx:          ctx,
		logger:       logger,
		notification: notificationManager,
	}
}

func (c *AuthController) parseAndValidate(w http.ResponseWriter, r *http.Request, req any) error {
	if err := c.validator.ValidateRequest(req); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		return err
	}

	if r.URL.Path == "/api/v1/auth/login" {
		return nil
	}

	return nil
}

func (c *AuthController) Notify(payloadType notification.NotificationPayloadType, user *shared_types.User, r *http.Request) {
	if r == nil {
		return
	}
	notificationData := notification.NotificationAuthenticationData{
		Email: user.Email,
		NotificationBaseData: notification.NotificationBaseData{
			IP:      r.RemoteAddr,
			Browser: r.UserAgent(),
		},
		UserName: user.Username,
	}

	payload := notification.NewNotificationPayload(
		payloadType,
		user.ID.String(),
		notificationData,
		notification.NotificationCategoryAuthentication,
	)

	c.notification.SendNotification(payload)
}
