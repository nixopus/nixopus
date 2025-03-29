package service

import (
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/utils"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// RefreshToken takes a refresh token as input and returns a new access token and refresh token. It
// first verifies the refresh token, then finds the associated user, creates a new access token,
// revokes the old refresh token and creates a new one, and finally returns the new access token,
// refresh token, the user's expiration time, and the user information.
func (c *AuthService) RefreshToken(refreshToken types.RefreshTokenRequest) (types.AuthResponse, error) {
	if refreshToken.RefreshToken == "" {
		return types.AuthResponse{}, types.ErrRefreshTokenIsRequired
	}

	c.logger.Log(logger.Info, "refreshing token", refreshToken.RefreshToken)
	token, err := c.storage.GetRefreshToken(refreshToken.RefreshToken)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrInvalidRefreshToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrInvalidRefreshToken
	}

	user, err := c.storage.FindUserByID(token.UserID.String())
	if err != nil {
		c.logger.Log(logger.Error, types.ErrUserNotFound.Error(), err.Error())
		return types.AuthResponse{}, types.ErrUserNotFound
	}

	accessToken, err := utils.CreateToken(user.Email, time.Minute*15)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateToken
	}

	c.storage.RevokeRefreshToken(token.Token)
	newRefreshToken, err := c.storage.CreateRefreshToken(user.ID)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateRefreshToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateRefreshToken
	}

	return types.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken.Token,
		ExpiresIn:    newRefreshToken.ExpiresAt.Unix(),
		User:         *user,
	}, nil
}
