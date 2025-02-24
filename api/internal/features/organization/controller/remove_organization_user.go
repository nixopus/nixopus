package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
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

	if err := c.service.RemoveUserFromOrganization(&user); err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "User removed from organization successfully", nil)
}
