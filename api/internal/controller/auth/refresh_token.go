package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
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
	err := json.NewDecoder(r.Body).Decode(&refreshRequest)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if refreshRequest.RefreshToken == "" {
		utils.SendErrorResponse(w, types.ErrRefreshTokenIsRequired.Error(), http.StatusBadRequest)
		return
	}

	refreshToken, err := storage.GetRefreshToken(c.app.Store.DB, refreshRequest.RefreshToken, c.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrInvalidRefreshToken.Error(), http.StatusUnauthorized)
		return
	}

	user, err := storage.FindUserByID(c.app.Store.DB, refreshToken.UserID.String(), c.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrUserNotFound.Error(), http.StatusNotFound)
		return
	}

	accessToken, err := createToken(user.Email, time.Minute*15)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToCreateToken.Error(), http.StatusInternalServerError)
		return
	}

	storage.RevokeRefreshToken(c.app.Store.DB, refreshRequest.RefreshToken, c.app.Ctx)
	newRefreshToken, err := storage.CreateRefreshToken(c.app.Store.DB, user.ID, c.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToCreateRefreshToken.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Token refreshed successfully", types.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken.Token,
		ExpiresIn:    15 * 60,
		User:         *user,
	})
}
