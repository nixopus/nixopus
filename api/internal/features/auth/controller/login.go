package auth

import (
	"net/http"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/utils"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (ar *AuthController) Login(c fuego.ContextWithBody[types.LoginRequest]) (*shared_types.Response, error) {
	loginRequest, err := c.Body()

	ar.logger.Log(logger.Info, "logging in user", loginRequest.Email)
	if err != nil {
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

		return &shared_types.Response{
			Status:  "2fa_required",
			Message: "Two-factor authentication required",
			Data: map[string]interface{}{
				"temp_token": tempToken,
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

	return &shared_types.Response{
		Status:  "success",
		Message: "User logged in successfully",
		Data:    response,
	}, nil
}
