package auth

import (
	"log"
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// ResetPassword handles HTTP requests to reset the user's password.
//
// The function expects a JSON body of type types.ChangePasswordRequest containing the user's old and new passwords.
//
// If the request body cannot be decoded, it responds with a 400 error.
// If the old or new password is empty, it responds with a 400 error.
// If the old password does not match the user's current password, it responds with a 401 error.
// If the user's password cannot be updated, it responds with a 500 error.
//
// On successful reset, it responds with a 200 status code and an empty JSON response.
func (c *AuthController) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var reset_password_request types.ChangePasswordRequest
	err := c.validator.ParseRequestBody(r, r.Body, &reset_password_request)

	if err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	err = c.validator.ValidateRequest(reset_password_request)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	userAny := r.Context().Value(shared_types.UserContextKey)
	user, ok := userAny.(*shared_types.User)

	err = c.service.ResetPassword(user, reset_password_request)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !ok {
		log.Println("Failed to get user from context")
		utils.SendErrorResponse(w, types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Password reset successfully", nil)
}

// GeneratePasswordResetLink handles HTTP requests to generate a password reset link for a user.
//
// It expects a user to be present in the context.
//
// If the user is not present in the context, it responds with a 500 error.
// If a token cannot be created, it responds with a 500 error.
// If the email with the reset link cannot be sent, it responds with a 500 error.
// If the user cannot be updated, it responds with a 500 error.
//
// On successful generation of the password reset link, it responds with a 200 status code and an empty JSON response.
func (c *AuthController) GeneratePasswordResetLink(w http.ResponseWriter, r *http.Request) {
	userAny := r.Context().Value(shared_types.UserContextKey)
	user, ok := userAny.(*shared_types.User)

	if !ok {
		log.Println("Failed to get user from context")
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
