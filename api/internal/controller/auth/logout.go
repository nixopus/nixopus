package auth

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
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

	err := json.NewDecoder(r.Body).Decode(&logoutRequest)
	if err != nil {
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if logoutRequest.RefreshToken != "" {
		err = storage.RevokeRefreshToken(c.app.Store.DB, logoutRequest.RefreshToken, c.app.Ctx)
		if err != nil {
			log.Printf("Failed to revoke refresh token: %v", err)
		}
	}

	utils.SendJSONResponse(w, "success", "Logged out successfully", nil)
}
