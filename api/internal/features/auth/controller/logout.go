package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *AuthController) Logout(s fuego.ContextWithBody[types.LogoutRequest]) (*shared_types.Response, error) {
	logoutRequest, err := s.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := s.Response(), s.Request()
	if !c.parseAndValidate(w, r, &logoutRequest) {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(w, r)
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	if err := c.service.Logout(logoutRequest.RefreshToken); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	c.Notify(notification.NotificationPayloadTypeLogout, user, r)

	return &shared_types.Response{
		Status:  "success",
		Message: "User logged out successfully",
		Data:    nil,
	}, nil
}
