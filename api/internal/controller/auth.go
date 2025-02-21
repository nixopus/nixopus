package controller

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/storage"
	"github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type AuthController struct {
	app *storage.App
}

// NewAuthController creates a new AuthController with the given App.
//
// This function creates a new AuthController with the given App and returns a pointer to it.
//
// The App passed to this function should be a valid App that has been created with storage.NewApp.
func NewAuthController(app *storage.App) *AuthController {
	return &AuthController{
		app: app,
	}
}

// emptyRegistrationRequest checks if the registration request is empty.
//
// This function takes a types.RegisterRequest as input and checks if any of the
// fields are empty. If any of the fields are empty, it will return true. Otherwise,
// it will return false.
func emptyRegistrationRequest(registration_request types.RegisterRequest) bool {
	return registration_request.Username == "" || registration_request.Email == "" || registration_request.Password == ""
}

// createToken generates a JWT token for the given email.
//
// This function creates a new JWT token using the HS256 signing method. The token
// includes the user's email and an expiration time set to 24 hours from the time
// of creation. It returns the signed token string or an error if the signing process
// fails.
func createToken(email string, duration time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"email": email,
			"exp":   time.Now().Add(duration).Unix(),
		})

	tokenString, err := token.SignedString(types.JWTSecretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// HashPassword generates a hashed version of the given password using bcrypt.
//
// This function takes a plaintext password as input and returns a hashed version
// of the password. It uses the bcrypt algorithm with the default cost for hashing.
// If the hashing process fails, it returns an error describing the failure.
func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashedPassword), nil
}

// Register handles HTTP requests to register a new user.
//
// It expects a JSON body of type types.RegisterRequest containing the user's
// username, email, and password.
//
// If the request body cannot be decoded, it responds with a 400 error.
// If any of the fields are empty, it responds with a 400 error.
// If the password is invalid, it responds with a 400 error.
// If the password hashing fails, it responds with a 500 error.
// If a user with the provided email already exists, it responds with a 400 error.
// If the user cannot be registered, it responds with a 500 error.
// If a token cannot be created, it responds with a 500 error.
//
// On successful registration, it responds with a 200 status code and a JSON
// response containing the authentication token and user information.
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	var registration_request types.RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&registration_request)
	if err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if emptyRegistrationRequest(registration_request) {
		utils.SendErrorResponse(w, types.ErrEmptyPassword.Error(), http.StatusBadRequest)
		return
	}

	var user types.User
	user = user.
		SetUserName(registration_request.Username).
		SetEmail(registration_request.Email).
		SetPassword(registration_request.Password)

	if err := user.IsValidPassword(user.Password); err != nil {
		utils.SendErrorResponse(w, types.ErrInvalidPassword.Error(), http.StatusBadRequest)
		return
	}

	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToHashPassword.Error(), http.StatusInternalServerError)
		return
	}

	user = user.SetPassword(hashedPassword)

	user = user.NewUser()

	if db_user, err := storage.FindUserByEmail(c.app.Store.DB, user.Email, c.app.Ctx); err == nil {
		if db_user.ID != uuid.Nil {
			utils.SendErrorResponse(w, types.ErrUserWithEmailAlreadyExists.Error(), http.StatusBadRequest)
			return
		}
	}

	err = storage.CreateUser(c.app.Store.DB, &user, c.app.Ctx)
	if err != nil {
		log.Println(err.Error())
		utils.SendErrorResponse(w, types.ErrFailedToRegisterUser.Error(), http.StatusInternalServerError)
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

	utils.SendJSONResponse(w, "success", "User registered successfully", types.AuthResponse{
		AccessToken:  accessToken,
		User:         user,
		RefreshToken: refreshToken.Token,
		ExpiresIn:    900,
	})
}

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
		utils.SendErrorResponse(w, "Refresh token is required", http.StatusBadRequest)
		return
	}

	refreshToken, err := storage.GetRefreshToken(c.app.Store.DB, refreshRequest.RefreshToken, c.app.Ctx)
	if err != nil {
		utils.SendErrorResponse(w, "Invalid refresh token", http.StatusUnauthorized)
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

// SendVerificationEmail handles HTTP requests to send a verification email to a user.
//
// It expects the user to be present in the context.
//
// If the user is not present in the context, it responds with a 500 error.
//
// On successful sending of the verification email, it responds with a 200 status code and a JSON response
// indicating successful verification email sending.
func (c *AuthController) SendVerificationEmail(w http.ResponseWriter, r *http.Request) {

}

func (c *AuthController) VerifyEmail(w http.ResponseWriter, r *http.Request) {

}