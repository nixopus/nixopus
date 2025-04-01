package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *OrganizationsController) CreateOrganization(f fuego.ContextWithBody[types.CreateOrganizationRequest]) (*shared_types.Response, error) {
	organization, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	c.logger.Log("Creating organization", organization.Name, organization.Description)

	w, r := f.Response(), f.Request()
	if err := c.validator.ValidateRequest(&organization); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	loggedInUser := utils.GetUser(w, r)
	if loggedInUser == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	createdOrganization, err := c.service.CreateOrganization(&organization)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	roles, err := c.role_service.GetRoleByName(shared_types.RoleAdmin)
	if err != nil {
		c.logger.Log(logger.Error, "failed to get role by name", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	if roles == nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	c.service.AddUserToOrganization(types.AddUserToOrganizationRequest{
		UserID:         loggedInUser.ID.String(),
		OrganizationID: createdOrganization.ID.String(),
		RoleId:         roles.ID.String(),
	})

	c.Notify(notification.NortificationPayloadTypeCreateOrganization, loggedInUser, r)

	return &shared_types.Response{
		Status:  "success",
		Message: "Organization created successfully",
		Data:    createdOrganization,
	}, nil
}
