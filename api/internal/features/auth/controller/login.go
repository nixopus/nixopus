package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
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
