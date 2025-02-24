package auth

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
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
	err := c.validator.ParseRequestBody(r, r.Body, &logoutRequest)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	err = c.validator.ValidateRequest(logoutRequest)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.service.Logout(logoutRequest.RefreshToken); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Logged out successfully", nil)
}
