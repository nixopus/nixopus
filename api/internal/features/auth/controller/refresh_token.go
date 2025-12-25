package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
)

func (c *AuthController) RefreshToken(s fuego.ContextWithBody[types.RefreshTokenRequest]) (*types.LoginResponse, error) {
	refreshRequest, err := s.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := s.Response(), s.Request()
	if err := c.parseAndValidate(w, r, &refreshRequest); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	accessTokenResponse, err := c.service.RefreshToken(refreshRequest)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.LoginResponse{
		Status:  "success",
		Message: "Access token refreshed",
		Data:    accessTokenResponse,
	}, nil
}
