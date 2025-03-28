package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *AuthController) SendVerificationEmail(fuego.ContextNoBody) (types.Response, error) {
	return types.Response{
			Status:  "success",
			Message: "Verification email sent",
			Data:    nil,
		}, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusOK,
		}
}

func (c *AuthController) VerifyEmail(fuego.ContextNoBody) (types.Response, error) {
	return types.Response{
			Status:  "success",
			Message: "Email verified",
			Data:    nil,
		}, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusOK,
		}
}
