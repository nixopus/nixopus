package auth

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// RefreshToken godoc
// @Summary Refresh token endpoint
// @Description Refreshes a users access token with a new one.
// @Tags auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param refresh body types.RefreshTokenRequest true "Refresh request"
// @Success 200 {object} types.AuthResponse "Success response with token"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 401 {object} types.Response "Unauthorized"
// @Router /auth/refresh-token [post]
func (c *AuthController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var refreshRequest types.RefreshTokenRequest
	err := c.validator.ParseRequestBody(r, r.Body, &refreshRequest)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	err = c.validator.ValidateRequest(refreshRequest)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	accessTokenResponse, err := c.service.RefreshToken(refreshRequest)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusUnauthorized)
		return
	}

	utils.SendJSONResponse(w, "success", "Token refreshed successfully", accessTokenResponse)
}
