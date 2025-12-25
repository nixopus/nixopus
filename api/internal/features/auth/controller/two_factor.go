package auth

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/auth/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *AuthController) SetupTwoFactor(ctx fuego.ContextNoBody) (*types.TwoFactorSetupResponseWrapper, error) {
	user := utils.GetUser(ctx.Response(), ctx.Request())
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    types.ErrFailedToGetUserFromContext,
			Status: http.StatusUnauthorized,
		}
	}

	response, err := c.service.SetupTwoFactor(user)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.TwoFactorSetupResponseWrapper{
		Status:  "success",
		Message: "Two-factor authentication setup successfully",
		Data:    response,
	}, nil
}

func (c *AuthController) VerifyTwoFactor(ctx fuego.ContextWithBody[types.TwoFactorVerifyRequest]) (*types.MessageResponse, error) {
	user := utils.GetUser(ctx.Response(), ctx.Request())
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    types.ErrFailedToGetUserFromContext,
			Status: http.StatusUnauthorized,
		}
	}

	request, err := ctx.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	if err := c.service.VerifyTwoFactor(user, request.Code); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "Two-factor authentication enabled successfully",
	}, nil
}

func (c *AuthController) DisableTwoFactor(ctx fuego.ContextNoBody) (*types.MessageResponse, error) {
	user := utils.GetUser(ctx.Response(), ctx.Request())
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    types.ErrFailedToGetUserFromContext,
			Status: http.StatusUnauthorized,
		}
	}

	if err := c.service.DisableTwoFactor(user); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "Two-factor authentication disabled successfully",
	}, nil
}

func (c *AuthController) TwoFactorLogin(ctx fuego.ContextWithBody[types.TwoFactorLoginRequest]) (*types.LoginResponse, error) {
	request, err := ctx.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	user, err := c.service.GetUserByEmail(request.Email)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusUnauthorized,
		}
	}

	if !user.TwoFactorEnabled {
		return nil, fuego.HTTPError{
			Err:    types.ErrInvalid2FACode,
			Status: http.StatusBadRequest,
		}
	}

	if err := c.service.VerifyTwoFactorCode(user, request.Code); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusUnauthorized,
		}
	}

	response, err := c.service.Login(request.Email, request.Password)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusUnauthorized,
		}
	}

	return &types.LoginResponse{
		Status:  "success",
		Message: "User logged in successfully",
		Data:    response,
	}, nil
}
