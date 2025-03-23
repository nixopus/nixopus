package auth

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
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

	if !c.parseAndValidate(w, r, &reset_password_request) {
		return
	}

	user := c.GetUser(w, r)
	if user == nil {
		return
	}

	err := c.service.ResetPassword(user, reset_password_request)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Notify(notification.NotificationPayloadTypePasswordReset, user, r)

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
// @Router /auth/request-password-reset [post]
func (c *AuthController) GeneratePasswordResetLink(w http.ResponseWriter, r *http.Request) {
	user := c.GetUser(w, r)
	if user == nil {
		return
	}
	err := c.service.GeneratePasswordResetLink(user)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Password reset link sent successfully", nil)
}
