package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *OrganizationsController) UpdateOrganization(f fuego.ContextWithBody[types.UpdateOrganizationRequest]) (*shared_types.Response, error) {
	_, r := f.Response(), f.Request()
	organization, err := f.Body()
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

	if err := c.service.UpdateOrganization(&organization); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	// c.Notify(notification.NotificationPayloadTypeUpdateOrganization, loggedInUser, r)

	return &shared_types.Response{
		Status:  "success",
		Message: "Organization updated successfully",
		Data:    nil,
	}, nil
}
