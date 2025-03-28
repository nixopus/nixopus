package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// AddUserToOrganization godoc
// @Summary Add a user to an organization
// @Description Adds a user to an organization
// @Tags organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param add_user_to_organization body types.AddUserToOrganizationRequest true "Add user to organization request"
// @Success 200 {object} types.Response "Success response with user"
// @Failure 400 {object} types.Response "Bad request"
// @Router /organization/add-user-to-organization [post]
func (c *OrganizationsController) AddUserToOrganization(w http.ResponseWriter, r *http.Request) {
	var user types.AddUserToOrganizationRequest

	if !c.parseAndValidate(w, r, &user) {
		return
	}

	loggedInUser := utils.GetUser(w, r)
	if loggedInUser == nil {
		return
	}

	err := c.service.AddUserToOrganization(user)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Notify(notification.NortificationPayloadTypeAddUserToOrganization, loggedInUser, r)

	utils.SendJSONResponse(w, "success", "User added to organization successfully", nil)
}
