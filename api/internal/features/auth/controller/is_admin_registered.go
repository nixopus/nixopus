package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
)

func (c *AuthController) IsAdminRegistered(s fuego.ContextNoBody) (*types.AdminRegisteredResponse, error) {
	isAdminRegistered, err := c.service.IsAdminRegistered()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.AdminRegisteredResponse{
		Status:  "success",
		Message: "Admin registration status",
		Data: types.AdminRegisteredData{
			AdminRegistered: isAdminRegistered,
		},
	}, nil
}
