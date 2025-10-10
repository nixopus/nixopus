package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *AuthController) ResetPassword(s fuego.ContextWithBody[types.ResetPasswordRequest]) (shared_types.Response, error) {
	reset_password_request, err := s.Body()
	if err != nil {
		return shared_types.Response{}, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := s.Response(), s.Request()
	if err := c.parseAndValidate(w, r, &reset_password_request); err != nil {
		return shared_types.Response{}, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	token := r.URL.Query().Get("token")
	if token == "" {
		return shared_types.Response{}, fuego.HTTPError{
			Err:    types.ErrInvalidResetToken,
			Status: http.StatusBadRequest,
		}
	}

	user, err := c.service.GetUserByResetToken(token)
	if err != nil {
		return shared_types.Response{}, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	err = c.service.ResetPassword(user, reset_password_request)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return shared_types.Response{}, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	c.Notify(notification.NotificationPayloadTypePasswordReset, user, r)

	return shared_types.Response{
		Status:  "success",
		Message: "Password reset successfully",
		Data:    nil,
	}, nil
}
