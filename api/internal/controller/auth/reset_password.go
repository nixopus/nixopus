package auth

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"golang.org/x/crypto/bcrypt"
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
	err := json.NewDecoder(r.Body).Decode(&reset_password_request)

	if err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	userAny := r.Context().Value(types.UserContextKey)
	user, ok := userAny.(*types.User)

	if !ok {
		log.Println("Failed to get user from context")
		utils.SendErrorResponse(w, types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return
	}

	if reset_password_request.NewPassword == "" || reset_password_request.OldPassword == "" {
		utils.SendErrorResponse(w, types.ErrEmptyPassword.Error(), http.StatusBadRequest)
		return
	}

	if reset_password_request.NewPassword == reset_password_request.OldPassword {
		utils.SendErrorResponse(w, types.ErrSamePassword.Error(), http.StatusBadRequest)
		return
	}

	user, err = storage.GetResetToken(c.app.Store.DB, user.ResetToken, c.app.Ctx)

	fmt.Printf("user: %v\n", user.ResetToken)

	if user.ResetToken == "" || err != nil {
		utils.SendErrorResponse(w, types.ErrInvalidResetToken.Error(), http.StatusBadRequest)
		return
	}

	jwtToken, err := jwt.Parse(user.ResetToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return types.JWTSecretKey, nil
	})

	if err != nil || !jwtToken.Valid {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrInvalidResetToken.Error(), http.StatusBadRequest)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reset_password_request.OldPassword)); err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrInvalidPassword.Error(), http.StatusUnauthorized)
		return
	}

	hashedPassword, err := HashPassword(reset_password_request.NewPassword)
	if err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToHashPassword.Error(), http.StatusInternalServerError)
		return
	}

	user.Password = hashedPassword

	err = storage.UpdateUser(c.app.Store.DB, user, c.app.Ctx)
	if err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToUpdateUser.Error(), http.StatusInternalServerError)
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
	userAny := r.Context().Value(types.UserContextKey)
	user, ok := userAny.(*types.User)

	if !ok {
		log.Println("Failed to get user from context")
		utils.SendErrorResponse(w, types.ErrFailedToGetUserFromContext.Error(), http.StatusInternalServerError)
		return
	}

	token, err := createToken(user.Email, time.Minute*5)
	if err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToCreateToken.Error(), http.StatusInternalServerError)
		return
	}

	user.ResetToken = token

	// handle sending email with reset link
	// err = utils.SendPasswordResetLinkEmail(user.Email, token)
	// if err != nil {
	// 	log.Println(err.Error())
	// 	utils.SendErrorResponse(w, types.ErrFailedToSendEmail.Error(), http.StatusInternalServerError)
	// 	return
	// }

	err = storage.UpdateUser(c.app.Store.DB, user, c.app.Ctx)
	if err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToUpdateUser.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Password reset link sent successfully", nil)
}