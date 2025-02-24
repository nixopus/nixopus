package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
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
	if err := c.validator.ParseRequestBody(r, r.Body, &user); err != nil {
		c.logger.Log(logger.Error, shared_types.ErrFailedToDecodeRequest.Error(), err.Error())
		utils.SendErrorResponse(w, shared_types.ErrFailedToDecodeRequest.Error(), http.StatusBadRequest)
		return
	}

	if err := c.validator.ValidateRequest(user); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.SendJSONResponse(w, "success", "User added to organization successfully", nil)
}
