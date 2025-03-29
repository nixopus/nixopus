package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *AuthController) Register(f fuego.ContextWithBody[types.RegisterRequest]) (*shared_types.Response, error) {
	w, _ := f.Response(), f.Request()
	registration_request, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	userResponse, err := c.service.Register(registration_request)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return nil, nil
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "User created successfully",
		Data:    userResponse,
	}, nil
}

func (c *AuthController) CreateUser(s fuego.ContextWithBody[types.RegisterRequest]) (*shared_types.Response, error) {
	registration_request, err := s.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	w, r := s.Response(), s.Request()
	if !c.parseAndValidate(w, r, &registration_request) {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
	}

	userResponse, err := c.service.Register(registration_request)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "User created successfully",
		Data:    userResponse,
	}, nil
}
