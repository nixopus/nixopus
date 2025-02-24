package service

import (
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// Register registers a new user and returns an authentication response.
//
// The function takes a types.RegisterRequest as input, which includes the user's
// username, email, and password. It first hashes the password and checks if a user
// with the provided email already exists. If a user exists, it returns an error.
// If not, it creates a new user in the database. It then creates a refresh token
// and an access token for the user.
//
// Returns a types.AuthResponse containing the access token, refresh token, expiration
// time, and user information. If any step fails, it returns an appropriate error.
func (c *AuthService) Register(registration_request types.RegisterRequest) (types.AuthResponse, error) {
	var user shared_types.User
	c.logger.Log(logger.Info, "registering user", registration_request.Email)
	hashedPassword, err := HashPassword(user.Password)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToHashPassword.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToHashPassword
	}

	user = shared_types.NewUser(registration_request.Email, hashedPassword, registration_request.Username, "")

	if db_user, err := c.storage.FindUserByEmail(registration_request.Email); err == nil {
		c.logger.Log(logger.Error, types.ErrUserWithEmailAlreadyExists.Error(), "")
		if db_user.ID != uuid.Nil {
			return types.AuthResponse{}, types.ErrUserWithEmailAlreadyExists
		}
	}

	err = c.storage.CreateUser(&user)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToRegisterUser.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToRegisterUser
	}

	refreshToken, err := c.storage.CreateRefreshToken(user.ID)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateRefreshToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateToken
	}

	accessToken, err := createToken(user.Email, time.Minute*15)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateAccessToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateToken
	}

	return types.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		ExpiresIn:    refreshToken.ExpiresAt.Unix(),
		User:         user,
	}, nil
}
