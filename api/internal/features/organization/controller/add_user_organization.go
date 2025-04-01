package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *OrganizationsController) AddUserToOrganization(f fuego.ContextWithBody[types.AddUserToOrganizationRequest]) (*shared_types.Response, error) {
	_, r := f.Response(), f.Request()
	user, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	loggedInUser := utils.GetUser(f.Response(), r)
	if loggedInUser == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	err = c.service.AddUserToOrganization(user)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	c.Notify(notification.NortificationPayloadTypeAddUserToOrganization, loggedInUser, r)

	return &shared_types.Response{
		Status:  "success",
		Message: "User added to organization successfully",
		Data:    nil,
	}, nil
}
