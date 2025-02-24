package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// GetOrganizations godoc
// @Summary Get all organizations
// @Description Retrieves all organizations.
// @Tags organization
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} types.Organization "Success response with organizations"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /organizations [get]
func (c *OrganizationsController) GetOrganizations(w http.ResponseWriter, r *http.Request) {
	organizations, err := c.service.GetOrganizations()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}
	utils.SendJSONResponse(w, "success", "Organizations fetched successfully", organizations)
}
