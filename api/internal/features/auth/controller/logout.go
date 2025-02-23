package auth

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// Logout handles HTTP requests to log out a user.
//
// It expects a JSON body of type types.LogoutRequest containing the user's
// refresh token.
//
// If the request body cannot be decoded, it responds with a 400 error.
// If the refresh token is provided, it attempts to revoke the token.
// If revoking the refresh token fails, it logs the error.
//
// On successful logout, it responds with a 200 status code and a JSON response
// indicating successful logout.
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	var logoutRequest types.LogoutRequest

	err := c.validator.ParseRequestBody(r, r.Body, &logoutRequest)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	err = c.validator.ValidateRequest(logoutRequest)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.service.Logout(logoutRequest.RefreshToken); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Logged out successfully", nil)
}
