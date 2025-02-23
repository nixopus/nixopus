package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"golang.org/x/crypto/bcrypt"
)

func (c *AuthService) ResetPassword(user *shared_types.User, reset_password_request types.ChangePasswordRequest) error {
	user, err := c.storage.GetResetToken(user.ResetToken)

	if user.ResetToken == "" || err != nil {
		return types.ErrInvalidResetToken
	}

	jwtToken, err := jwt.Parse(user.ResetToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return shared_types.JWTSecretKey, nil
	})

	if err != nil || !jwtToken.Valid {
		return types.ErrInvalidResetToken
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(reset_password_request.OldPassword)); err != nil {
		return types.ErrInvalidPassword
	}

	hashedPassword, err := HashPassword(reset_password_request.NewPassword)
	if err != nil {
		return types.ErrFailedToHashPassword
	}

	user.Password = hashedPassword

	err = c.storage.UpdateUser(user)
	if err != nil {
		return types.ErrFailedToUpdateUser
	}

	return nil
}

func (c *AuthService) GeneratePasswordResetLink(user *shared_types.User) error {
	token, err := createToken(user.Email, time.Minute*5)
	if err != nil {
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
		return types.ErrFailedToUpdateUser
	}

	return nil
}
