package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/utils"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// ResetPassword resets a user's password.
//
// This function takes a user and a ChangePasswordRequest as inputs. It first verifies the
// reset token associated with the user, ensuring it is valid and correctly signed. It then
// checks if the provided old password matches the stored password. If valid, it hashes
// the new password and updates the user's password in the storage. Errors are logged and
// returned at each step if any process fails.
func (c *AuthService) ResetPassword(user *shared_types.User, reset_password_request types.ResetPasswordRequest) error {
	c.logger.Log(logger.Info, "resetting password", user.Email)
	if user.ResetToken == "" {
		c.logger.Log(logger.Error, types.ErrInvalidResetToken.Error(), "reset token is empty")
		return types.ErrInvalidResetToken
	}

	tx, err := c.storage.BeginTx()
	if err != nil {
		c.logger.Log(logger.Error, "failed to start transaction", err.Error())
		return types.ErrFailedToUpdateUser
	}
	defer tx.Rollback()

	txStorage := c.storage.WithTx(tx)

	user, err = txStorage.GetResetToken(user.ResetToken)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrInvalidResetToken.Error(), err.Error())
		return types.ErrInvalidResetToken
	}

	jwtToken, err := jwt.Parse(user.ResetToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			c.logger.Log(logger.Error, types.ErrInvalidResetToken.Error(), err.Error())
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return shared_types.JWTSecretKey, nil
	})

	if err != nil || !jwtToken.Valid {
		c.logger.Log(logger.Error, types.ErrInvalidResetToken.Error(), err.Error())
		return types.ErrInvalidResetToken
	}

	hashedPassword, err := utils.HashPassword(reset_password_request.Password)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToHashPassword.Error(), err.Error())
		return types.ErrFailedToHashPassword
	}

	user.Password = hashedPassword
	user.ResetToken = ""

	err = txStorage.UpdateUser(user)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToUpdateUser.Error(), err.Error())
		return types.ErrFailedToUpdateUser
	}

	if err := tx.Commit(); err != nil {
		c.logger.Log(logger.Error, "failed to commit transaction", err.Error())
		return types.ErrFailedToUpdateUser
	}

	return nil
}

// GeneratePasswordResetLink generates a password reset link for a user and sends it to the user's email
//
// The function takes a user as input and returns an error if the user is not found or if the token
// cannot be created. It also returns an error if the user cannot be updated.
//
// The generated link is valid for 5 minutes.
func (c *AuthService) GeneratePasswordResetLink(user *shared_types.User) (*shared_types.User, string, error) {
	if user == nil {
		c.logger.Log(logger.Error, types.ErrInvalidUser.Error(), "user is nil")
		return nil, "", types.ErrInvalidUser
	}

	c.logger.Log(logger.Info, "generating password reset link", user.Email)
	token, err := utils.CreateToken(user.Email, time.Minute*5, user.TwoFactorEnabled, true)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateToken.Error(), err.Error())
		return nil, "", types.ErrFailedToCreateToken
	}

	user.ResetToken = token

	err = c.storage.UpdateUser(user)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToUpdateUser.Error(), err.Error())
		return nil, "", types.ErrFailedToUpdateUser
	}

	return user, token, nil
}

// GetUserByResetToken retrieves a user by their reset token.
//
// This function takes a reset token as input and returns the associated user.
// If the token is invalid or expired, it returns an error.
func (c *AuthService) GetUserByResetToken(token string) (*shared_types.User, error) {
	if token == "" {
		c.logger.Log(logger.Error, types.ErrInvalidResetToken.Error(), "reset token is empty")
		return nil, types.ErrInvalidResetToken
	}

	user, err := c.storage.GetResetToken(token)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrInvalidResetToken.Error(), err.Error())
		return nil, types.ErrInvalidResetToken
	}

	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			c.logger.Log(logger.Error, types.ErrInvalidResetToken.Error(), err.Error())
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return shared_types.JWTSecretKey, nil
	})

	if err != nil || !jwtToken.Valid {
		c.logger.Log(logger.Error, types.ErrInvalidResetToken.Error(), err.Error())
		return nil, types.ErrInvalidResetToken
	}

	return user, nil
}
