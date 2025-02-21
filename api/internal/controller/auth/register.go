package auth

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

	fmt.Println(registration_request)

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
