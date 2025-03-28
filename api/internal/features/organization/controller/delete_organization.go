package controller

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// DeleteOrganization godoc
// @Summary Delete an organization
// @Description Deletes an organization with the given id.
// @Tags organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param delete_organization body types.DeleteOrganizationRequest true "Delete organization request"
// @Success 200 {object} types.Response "Success response"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /organization/delete [post]
func (c *OrganizationsController) DeleteOrganization(w http.ResponseWriter, r *http.Request) {
	var organization types.DeleteOrganizationRequest

	loggedInUser := utils.GetUser(w, r)
	if loggedInUser == nil {
		return
	}

	if !c.parseAndValidate(w, r, &organization) {
		return
	}

	organizationID, err := uuid.Parse(organization.ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := c.service.DeleteOrganization(organizationID); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Notify(notification.NotificationPayloadTypeDeleteOrganization, loggedInUser, r)

	utils.SendJSONResponse(w, "success", "Organization deleted successfully", nil)
}
