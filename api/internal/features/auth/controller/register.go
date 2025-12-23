package auth

import (
	"errors"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *AuthController) Register(f fuego.ContextWithBody[types.RegisterRequest]) (*shared_types.Response, error) {
	w, r := f.Response(), f.Request()
	registration_request, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	if registration_request.Organization != "" {
		return nil, fuego.HTTPError{
			Err:    errors.New("organization is not supported in admin registration"),
			Status: http.StatusBadRequest,
		}
	}

	if registration_request.Type != shared_types.UserTypeAdmin {
		return nil, fuego.HTTPError{
			Err:    errors.New("type must be admin"),
			Status: http.StatusBadRequest,
		}
	}

	if err := c.parseAndValidate(w, r, &registration_request); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	adminRegistered, err := c.service.IsAdminRegistered()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	if adminRegistered {
		return nil, fuego.HTTPError{
			Err:    errors.New("admin already registered"),
			Status: http.StatusBadRequest,
		}
	}

	userResponse, err := c.service.Register(registration_request, shared_types.UserTypeAdmin)
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
	if err := c.parseAndValidate(w, r, &registration_request); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	userResponse, err := c.service.Register(registration_request, shared_types.UserTypeUser)
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
