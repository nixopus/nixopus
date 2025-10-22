package service

import (
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/utils"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"golang.org/x/crypto/bcrypt"
)

// Deprecated: Use SupertokensLogin instead
func (u *AuthService) Login(email string, password string) (types.AuthResponse, error) {
	u.logger.Log(logger.Info, "logging in user", email)

	tx, err := u.storage.BeginTx()
	if err != nil {
		u.logger.Log(logger.Error, "failed to start transaction", err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateToken
	}
	defer tx.Rollback()

	txStorage := u.storage.WithTx(tx)

	user, err := txStorage.FindUserByEmail(email)
	if err != nil {
		u.logger.Log(logger.Error, types.ErrUserNotFound.Error(), err.Error())
		return types.AuthResponse{}, types.ErrUserNotFound
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		u.logger.Log(logger.Error, types.ErrInvalidPassword.Error(), err.Error())
		return types.AuthResponse{}, types.ErrInvalidPassword
	}

	refreshToken, err := txStorage.CreateRefreshToken(user.ID)
	if err != nil {
		u.logger.Log(logger.Error, types.ErrFailedToCreateRefreshToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateRefreshToken
	}

	accessToken, err := utils.CreateToken(user.Email, time.Hour*24*7, user.TwoFactorEnabled, true)
	if err != nil {
		u.logger.Log(logger.Error, types.ErrFailedToCreateAccessToken.Error(), err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateAccessToken
	}

	if err := tx.Commit(); err != nil {
		u.logger.Log(logger.Error, "failed to commit transaction", err.Error())
		return types.AuthResponse{}, types.ErrFailedToCreateToken
	}

	return types.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken.Token,
		ExpiresIn:    7 * 24 * 60 * 60,
		User:         *user,
	}, nil
}
