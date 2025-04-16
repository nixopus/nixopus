package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

type UpdateUserRoleRequest struct {
	UserID         string `json:"user_id"`
	OrganizationID string `json:"organization_id"`
	RoleName       string `json:"role_name"`
}

func (c *OrganizationsController) UpdateUserRole(f fuego.ContextWithBody[UpdateUserRoleRequest]) (*shared_types.Response, error) {
	_, r := f.Response(), f.Request()
	req, err := f.Body()
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

	if err := c.service.UpdateUserRole(req.UserID, req.OrganizationID, req.RoleName); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	org, err := c.service.GetOrganization(req.OrganizationID)
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

	c.Notify(notification.NotificationPayloadTypeUpdateUserRole, loggedInUser, r, notification.UpdateUserRoleData{
		OrganizationName: org.Name,
		UserName:         userDetails.Username,
		NewRole:          req.RoleName,
	})

	return &shared_types.Response{
		Status:  "success",
		Message: "User role updated successfully",
		Data:    nil,
	}, nil
}
