package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/notification"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// CreateOrganization godoc
// @Summary Create a new organization
// @Description Creates a new organization in the application.
// @Tags organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param create_organization body types.CreateOrganizationRequest true "Create organization request"
// @Success 200 {object} types.Response "Success response with organization"
// @Failure 400 {object} types.Response "Bad request"
// @Router /organization/create [post]
func (c *OrganizationsController) CreateOrganization(w http.ResponseWriter, r *http.Request) {
	var organization types.CreateOrganizationRequest

	loggedInUser := c.GetUser(w, r)
	if loggedInUser == nil {
		return
	}

	if !c.parseAndValidate(w, r, &organization) {
		return
	}

	createdOrganization, err := c.service.CreateOrganization(&organization)

	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	c.Notify(notification.NortificationPayloadTypeCreateOrganization, loggedInUser, r)

	utils.SendJSONResponse(w, "success", "Organization created successfully", createdOrganization)
}
