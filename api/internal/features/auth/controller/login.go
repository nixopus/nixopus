package auth

import (
	"net/http"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/utils"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
)

func (ar *AuthController) Login(c fuego.ContextWithBody[types.LoginRequest]) (*types.LoginResponse, error) {
	loginRequest, err := c.Body()

	ar.logger.Log(logger.Info, "logging in user", loginRequest.Email)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	if err := ar.validator.ValidateRequest(&loginRequest); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	user, err := ar.service.GetUserByEmail(loginRequest.Email)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusUnauthorized,
		}
	}

	if user.TwoFactorEnabled {
		tempToken, err := utils.CreateToken(user.Email, 2*time.Minute, true, false)
		if err != nil {
			return nil, fuego.HTTPError{
				Err:    err,
				Status: http.StatusInternalServerError,
			}
		}

		return &types.LoginResponse{
			Status:  "2fa_required",
			Message: "Two-factor authentication required",
			Data: types.AuthResponse{
				AccessToken: tempToken,
			},
		}, nil
	}

	response, err := ar.service.Login(loginRequest.Email, loginRequest.Password)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusUnauthorized,
		}
	}

	ar.Notify(notification.NotificationPayloadTypeLogin, &response.User, c.Request())

	return &types.LoginResponse{
		Status:  "success",
		Message: "User logged in successfully",
		Data:    response,
	}, nil
}
