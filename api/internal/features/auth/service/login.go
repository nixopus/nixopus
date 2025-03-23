package service

import (
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"golang.org/x/crypto/bcrypt"
	"time"
)

// Login authenticates a user and returns an authentication token and user information.
//
// The function takes in an email and password as input and returns an error if the user
// does not exist or if the password is invalid. It also returns an error if a token
// cannot be created.
//
// The returned types.AuthResponse contains an authentication token, a refresh token,
// the user's expiration time, and the user information.
func (u *AuthService) Login(email string, password string) (types.AuthResponse, error) {
	u.logger.Log(logger.Info, "logging in user", email)
	user, err := u.storage.FindUserByEmail(email)
	if err != nil {
		u.logger.Log(logger.Error, types.ErrUserNotFound.Error(), err.Error())
		return types.AuthResponse{}, types.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		u.logger.Log(logger.Error, types.ErrInvalidPassword.Error(), err.Error())
		return types.AuthResponse{}, types.ErrInvalidPassword
	}

	refreshToken, err := u.storage.CreateRefreshToken(user.ID)
	if err != nil {
		u.logger.Log(logger.Error, types.ErrFailedToCreateRefreshToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateRefreshToken
	}

	accessToken, err := CreateToken(user.Email, time.Minute*15)
	if err != nil {
		u.logger.Log(logger.Error, types.ErrFailedToCreateAccessToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateAccessToken
	}

	return types.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		ExpiresIn:    900,
		User:         *user,
	}, nil
}
