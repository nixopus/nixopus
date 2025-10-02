package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/validation"
	"github.com/supertokens/supertokens-golang/recipe/passwordless"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *OrganizationsController) SendInvite(f fuego.ContextWithBody[types.InviteSendRequest]) (*shared_types.Response, error) {
	request, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	validator := validation.NewValidator(nil)
	if err := validator.ValidateRequest(&request); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	userContext := &map[string]interface{}{
		"email":           request.Email,
		"organization_id": request.OrganizationID,
		"role":            request.Role,
	}

	response, err := passwordless.CreateCodeWithEmail("public", request.Email, nil, userContext)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to create passwordless code", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	if response.OK == nil {
		return &shared_types.Response{
			Status:  "error",
			Message: "Failed to create magic link",
		}, nil
	}

	c.logger.Log(logger.Info, "Organization invite sent", request.Email)

	return &shared_types.Response{
		Status:  "success",
		Message: "Invite sent successfully",
		Data: map[string]interface{}{
			"email":           request.Email,
			"organization_id": request.OrganizationID,
			"role":            request.Role,
		},
	}, nil
}

func (c *OrganizationsController) ResendInvite(f fuego.ContextWithBody[types.InviteResendRequest]) (*shared_types.Response, error) {
	request, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	validator := validation.NewValidator(nil)
	if err := validator.ValidateRequest(&request); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	userContext := &map[string]interface{}{
		"email":           request.Email,
		"organization_id": request.OrganizationID,
		"role":            request.Role,
	}

	response, err := passwordless.CreateCodeWithEmail("public", request.Email, nil, userContext)
	if err != nil {
		c.logger.Log(logger.Error, "Failed to resend passwordless code", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	if response.OK == nil {
		return &shared_types.Response{
			Status:  "error",
			Message: "Failed to resend magic link",
		}, nil
	}

	c.logger.Log(logger.Info, "Organization invite resent", request.Email)

	return &shared_types.Response{
		Status:  "success",
		Message: "Invite resent successfully",
		Data: map[string]interface{}{
			"email":           request.Email,
			"organization_id": request.OrganizationID,
			"role":            request.Role,
		},
	}, nil
}
