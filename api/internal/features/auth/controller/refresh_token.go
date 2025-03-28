package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *AuthController) RefreshToken(s fuego.ContextWithBody[types.RefreshTokenRequest]) (shared_types.Response, error) {
	refreshRequest, err := s.Body()
	if err != nil {
		return shared_types.Response{}, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := s.Response(), s.Request()
	if !c.parseAndValidate(w, r, &refreshRequest) {
		return shared_types.Response{}, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	accessTokenResponse, err := c.service.RefreshToken(refreshRequest)
	if err != nil {
		return shared_types.Response{}, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return shared_types.Response{
		Status:  "success",
		Message: "Access token refreshed",
		Data:    accessTokenResponse,
	}, nil
}
