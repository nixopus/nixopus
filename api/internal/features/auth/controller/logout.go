package auth

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// Logout godoc
// @Summary Logout user endpoint
// @Description Logs out a user by revoking the refresh token.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param logout body types.LogoutRequest true "Logout request"
// @Success 200 {object} types.Response "Success response"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /auth/logout [post]
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	var logoutRequest types.LogoutRequest
	if !c.parseAndValidate(w, r, &logoutRequest) {
		return
	}

	user := c.GetUser(w, r)
	if user == nil {
		return
	}

	if err := c.service.Logout(logoutRequest.RefreshToken); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Notify(notification.NotificationPayloadTypeLogout, user, r)

	utils.SendJSONResponse(w, "success", "Logged out successfully", nil)
}
