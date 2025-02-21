package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// Login handles HTTP requests to authenticate a user and provide a token.
//
// It expects a JSON body of type types.LoginRequest containing the user's email and password.
//
// If the request body cannot be decoded, it responds with a 400 error.
// If the email or password is missing, it responds with a 400 error.
// If the user is not found, it responds with a 404 error.
// If the password is incorrect, it responds with a 401 error.
// If a token cannot be created, it responds with a 500 error.
//
// On successful authentication, it responds with a 200 status code and a JSON response
// containing the authentication token and user information.
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	var login_request types.LoginRequest
	err := json.NewDecoder(r.Body).Decode(&login_request)
	if err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if login_request.Email == "" || login_request.Password == "" {
		utils.SendErrorResponse(w, types.ErrEmptyPassword.Error(), http.StatusBadRequest)
		return
	}

	user, err := storage.FindUserByEmail(c.app.Store.DB, login_request.Email, c.app.Ctx)
	if err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrUserNotFound.Error(), http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(login_request.Password)); err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrInvalidPassword.Error(), http.StatusUnauthorized)
		return
	}

	refreshToken, err := storage.CreateRefreshToken(c.app.Store.DB, user.ID, c.app.Ctx)
	if err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToCreateToken.Error(), http.StatusInternalServerError)
		return
	}

	accessToken, err := createToken(user.Email, time.Minute*15)
	if err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToCreateToken.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "User logged in successfully", types.AuthResponse{
		AccessToken:  accessToken,
		User:         *user,
		RefreshToken: refreshToken.Token,
		ExpiresIn:    900,
	})
}
