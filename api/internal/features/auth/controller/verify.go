package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *AuthController) SendVerificationEmail(ctx fuego.ContextNoBody) (types.Response, error) {
	w, r := ctx.Response(), ctx.Request()
	user := utils.GetUser(w, r)
	if user == nil {
		return types.Response{
				Status:  "error",
				Message: "User not found",
				Data:    nil,
			}, fuego.HTTPError{
				Err:    nil,
				Status: http.StatusUnauthorized,
			}
	}

	token, err := c.service.GenerateVerificationToken(user.ID.String())
	if err != nil {
		return types.Response{
				Status:  "error",
				Message: "Failed to generate verification token",
				Data:    nil,
			}, fuego.HTTPError{
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

	return types.Response{
			Status:  "success",
			Message: "Verification email sent successfully",
			Data:    nil,
		}, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusOK,
		}
}

func (c *AuthController) VerifyEmail(ctx fuego.ContextNoBody) (types.Response, error) {
	token := ctx.QueryParam("token")
	if token == "" {
		return types.Response{
				Status:  "error",
				Message: "Verification token is required",
				Data:    nil,
			}, fuego.HTTPError{
				Err:    nil,
				Status: http.StatusBadRequest,
				Detail: "Verification token is required",
			}
	}

	userID, err := c.service.VerifyToken(token)
	if err != nil {
		return types.Response{
				Status:  "error",
				Message: "Invalid or expired verification token",
				Data:    nil,
			}, fuego.HTTPError{
				Err:    err,
				Status: http.StatusBadRequest,
				Detail: "Invalid or expired verification token",
			}
	}

	err = c.service.MarkEmailAsVerified(userID)
	if err != nil {
		return types.Response{
				Status:  "error",
				Message: "Failed to update user status",
				Data:    nil,
			}, fuego.HTTPError{
				Err:    err,
				Status: http.StatusInternalServerError,
			}
	}

	return types.Response{
			Status:  "success",
			Message: "Email verified successfully",
			Data:    nil,
		}, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusOK,
		}
}
