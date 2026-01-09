package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// TODO: Here we need to make sure when a user is removed from an organization, if no organization is left for the user, we should remove the user from the system.
func (c *OrganizationsController) RemoveUserFromOrganization(f fuego.ContextWithBody[types.RemoveUserFromOrganizationRequest]) (*types.MessageResponse, error) {
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

	if err := c.validator.ValidateRequest(&user); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	// Get organization details before removal for notification
	organization, err := c.service.GetOrganization(user.OrganizationID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	// Get user details - we'll fetch from the service's internal storage
	// Since the service already fetches this, we'll get it after removal
	// by querying the organization users or we can modify service to return it
	// For now, let's get it from the request context or fetch after removal

	// Note: The service already fetches the user, but we need it before removal
	// We'll need to modify the approach - get user info from organization users list
	orgUsers, err := c.service.GetOrganizationUsersWithRoles(user.OrganizationID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	var removedUser *shared_types.User
	for _, orgUser := range orgUsers {
		if orgUser.User.ID.String() == user.UserID {
			removedUser = orgUser.User
			break
		}
	}

	if removedUser == nil {
		return nil, fuego.HTTPError{
			Err:    types.ErrUserNotInOrganization,
			Status: http.StatusBadRequest,
		}
	}

	if err := c.service.RemoveUserFromOrganization(&user); err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	// Send notification to organization admins about user removal
	notificationData := notification.RemoveUserFromOrganizationData{
		NotificationBaseData: notification.NotificationBaseData{
			IP:      r.RemoteAddr,
			Browser: r.UserAgent(),
		},
		OrganizationName: organization.Name,
		UserName:         removedUser.Username,
		UserEmail:        removedUser.Email,
	}

	c.Notify(notification.NotificationPayloadTypeRemoveUserFromOrganization, loggedInUser, r, notificationData)

	return &types.MessageResponse{
		Status:  "success",
		Message: "User removed from organization successfully",
	}, nil
}
