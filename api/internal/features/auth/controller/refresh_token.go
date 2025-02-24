package auth

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// RefreshToken handles HTTP requests to refresh the user's access token.
//
// It expects a JSON body of type types.RefreshTokenRequest containing the user's
// refresh token.
//
// If the request body cannot be decoded, it responds with a 400 error.
// If the refresh token is empty, it responds with a 400 error.
// If the refresh token is invalid or expired, it responds with a 401 error.
// If the user is not found, it responds with a 404 error.
// If the access token cannot be created, it responds with a 500 error.
// If the refresh token cannot be revoked or a new one cannot be created, it responds with a 500 error.
//
// On successful refresh, it responds with a 200 status code and a JSON response
// containing the new access token, refresh token, and user information.
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
