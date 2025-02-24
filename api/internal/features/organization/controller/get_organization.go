package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/organization/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// GetOrganization godoc
// @Summary Get an organization
// @Description Retrieves an organization by its ID.
// @Tags organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Organization ID"
// @Success 200 {object} types.Organization "Success response with organization"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /organization/{id} [get]
func (c *OrganizationsController) GetOrganization(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if err := c.validator.ValidateID(id, types.ErrMissingOrganizationID); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}
	organization, err := c.service.GetOrganization(id)
	if err != nil {
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Organization fetched successfully", organization)
}
