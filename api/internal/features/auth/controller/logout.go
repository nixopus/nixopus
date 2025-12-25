package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *AuthController) Logout(s fuego.ContextWithBody[types.LogoutRequest]) (*types.MessageResponse, error) {
	c.logger.Log(logger.Info, "Logout request received", "")
	logoutRequest, err := s.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	c.logger.Log(logger.Info, "Logout request parsed", "")

	w, r := s.Response(), s.Request()
	if err := c.parseAndValidate(w, r, &logoutRequest); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	c.logger.Log(logger.Info, "Logout request validated", "")

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	c.logger.Log(logger.Info, "User found", "")

	if err := c.service.Logout(logoutRequest.RefreshToken); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	c.logger.Log(logger.Info, "Logout successful", "")

	c.Notify(notification.NotificationPayloadTypeLogout, user, r)

	c.logger.Log(logger.Info, "Logout notification sent", "")

	return &types.MessageResponse{
		Status:  "success",
		Message: "User logged out successfully",
	}, nil
}
