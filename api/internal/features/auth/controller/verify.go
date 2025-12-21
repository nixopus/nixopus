package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *AuthController) SendVerificationEmail(ctx fuego.ContextNoBody) (*types.MessageResponse, error) {
	w, r := ctx.Response(), ctx.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    types.ErrUserNotFound,
			Status: http.StatusUnauthorized,
		}
	}

	token, err := c.service.GenerateVerificationToken(user.ID.String())
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	notificationData := notification.NotificationVerificationEmailData{
		Email: user.Email,
		Token: token,
		NotificationBaseData: notification.NotificationBaseData{
			IP:      r.RemoteAddr,
			Browser: r.UserAgent(),
		},
	}

	payload := notification.NewNotificationPayload(
		notification.NotificationPayloadTypeVerificationEmail,
		user.ID.String(),
		notificationData,
		notification.NotificationCategoryAuthentication,
	)

	c.notification.SendNotification(payload)

	return &types.MessageResponse{
		Status:  "success",
		Message: "Verification email sent successfully",
	}, nil
}

func (c *AuthController) VerifyEmail(ctx fuego.ContextNoBody) (*types.MessageResponse, error) {
	token := ctx.QueryParam("token")
	if token == "" {
		return nil, fuego.HTTPError{
			Err:    types.ErrMissingRequiredFields,
			Status: http.StatusBadRequest,
		}
	}

	userID, err := c.service.VerifyToken(token)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	err = c.service.MarkEmailAsVerified(userID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "Email verified successfully",
	}, nil
}
