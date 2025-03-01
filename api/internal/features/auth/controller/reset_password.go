package auth

import (
	"log"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// ResetPassword godoc
// @Summary Reset password endpoint
// @Description Resets the user's password using the provided information.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param resetPassword body types.ChangePasswordRequest true "Reset password request"
// @Success 200 {object} types.Response "Password reset successfully"
// @Failure 400 {object} types.Response "Failed to decode or validate request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /auth/reset-password [post]
func (c *AuthController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var reset_password_request types.ChangePasswordRequest
	err := c.validator.ParseRequestBody(r, r.Body, &reset_password_request)

	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	err = c.validator.ValidateRequest(reset_password_request)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	userAny := r.Context().Value(shared_types.UserContextKey)
	user, ok := userAny.(*shared_types.User)

	if !ok {
		log.Println("Failed to get user from context")
		utils.SendErrorResponse(w, types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return
	}

	err = c.service.ResetPassword(user, reset_password_request)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.notification.SendNotification(notification.NewNotificationPayload(
		notification.NotificationPayloadTypePasswordReset,
		user.ID.String(),
		notification.NotificationAuthenticationData{
			Email: user.Email,
			NotificationBaseData: notification.NotificationBaseData{
				IP:      r.RemoteAddr,
				Browser: r.UserAgent(),
			},
			UserName: user.Username,
		},
		notification.NotificationCategoryAuthentication,
	))

	utils.SendJSONResponse(w, "success", "Password reset successfully", nil)
}

// GeneratePasswordResetLink godoc
// @Summary Generates a password reset link for a user
// @Description Generates a password reset link for a user and sends it to the user's email
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.Response "Password reset link sent successfully"
// @Failure 400 {object} types.Response "Failed to decode or validate request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /auth/generate-password-reset-link [post]
func (c *AuthController) GeneratePasswordResetLink(w http.ResponseWriter, r *http.Request) {
	userAny := r.Context().Value(shared_types.UserContextKey)
	user, ok := userAny.(*shared_types.User)

	if !ok {
		c.logger.Log(logger.Error, types.ErrFailedToGetUserFromContext.Error(), types.ErrFailedToGetUserFromContext.Error())
		utils.SendErrorResponse(w, types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return
	}

	err := c.service.GeneratePasswordResetLink(user)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Password reset link sent successfully", nil)
}
