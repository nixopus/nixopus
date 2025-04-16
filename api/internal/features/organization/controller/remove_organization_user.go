package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *OrganizationsController) RemoveUserFromOrganization(f fuego.ContextWithBody[types.RemoveUserFromOrganizationRequest]) (*shared_types.Response, error) {
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

	if err := c.service.RemoveUserFromOrganization(&user); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	org, err := c.service.GetOrganization(user.OrganizationID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	userDetails := utils.GetUser(f.Response(), r)
	if userDetails == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	c.Notify(notification.NotificationPayloadTypeRemoveUserFromOrganization, loggedInUser, r, notification.RemoveUserFromOrganizationData{
		NotificationBaseData: notification.NotificationBaseData{
			IP:      r.RemoteAddr,
			Browser: r.UserAgent(),
		},
		OrganizationName: org.Name,
		UserName:         userDetails.Username,
		UserEmail:        userDetails.Email,
	})

	return &shared_types.Response{
		Status:  "success",
		Message: "User removed from organization successfully",
		Data:    nil,
	}, nil
}
