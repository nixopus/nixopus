package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// RemoveUserFromOrganization godoc
// @Summary Remove a user from an organization
// @Description Removes a user from an organization
// @Tags organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param remove_user_from_organization body types.RemoveUserFromOrganizationRequest true "Remove user from organization request"
// @Success 200 {object} types.Response "Success response"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /organization/remove-user-from-organization [post]
func (c *OrganizationsController) RemoveUserFromOrganization(w http.ResponseWriter, r *http.Request) {
	var user types.RemoveUserFromOrganizationRequest

	loggedInUser := c.GetUser(w, r)
	if loggedInUser == nil {
		return
	}

	if !c.parseAndValidate(w, r, &user) {
		return
	}

	if err := c.service.RemoveUserFromOrganization(&user); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Notify(notification.NortificationPayloadTypeRemoveUserFromOrganization, loggedInUser, r)

	utils.SendJSONResponse(w, "success", "User removed from organization successfully", nil)
}
