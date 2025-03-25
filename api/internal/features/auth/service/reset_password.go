package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"golang.org/x/crypto/bcrypt"
)

// ResetPassword resets a user's password.
//
// This function takes a user and a ChangePasswordRequest as inputs. It first verifies the
// reset token associated with the user, ensuring it is valid and correctly signed. It then
// checks if the provided old password matches the stored password. If valid, it hashes
// the new password and updates the user's password in the storage. Errors are logged and
// returned at each step if any process fails.
func (c *AuthService) ResetPassword(user *shared_types.User, reset_password_request types.ChangePasswordRequest) error {
	c.logger.Log(logger.Info, "resetting password", user.Email)
	user, err := c.storage.GetResetToken(user.ResetToken)

	if user.ResetToken == "" || err != nil {
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

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reset_password_request.OldPassword)); err != nil {
		c.logger.Log(logger.Error, types.ErrInvalidPassword.Error(), err.Error())
		return types.ErrInvalidPassword
	}

	hashedPassword, err := HashPassword(reset_password_request.NewPassword)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToHashPassword.Error(), err.Error())
		return types.ErrFailedToHashPassword
	}

	user.Password = hashedPassword

	err = c.storage.UpdateUser(user)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToUpdateUser.Error(), err.Error())
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
func (c *AuthService) GeneratePasswordResetLink(user *shared_types.User) error {
	c.logger.Log(logger.Info, "generating password reset link", user.Email)
	token, err := CreateToken(user.Email, time.Minute*5)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToCreateToken.Error(), err.Error())
		return types.ErrFailedToCreateToken
	}

	user.ResetToken = token

	// handle sending email with reset link
	// err = utils.SendPasswordResetLinkEmail(user.Email, token)
	// if err != nil {
	// 	log.Println(err.Error())
	// 	utils.SendErrorResponse(w, types.ErrFailedToSendEmail.Error(), http.StatusInternalServerError)
	// 	return
	// }

	err = c.storage.UpdateUser(user)
	if err != nil {
		c.logger.Log(logger.Error, types.ErrFailedToUpdateUser.Error(), err.Error())
		return types.ErrFailedToUpdateUser
	}

	return nil
}
