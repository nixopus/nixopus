package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// @Summary Delete a domain
// @Description Deletes a domain by its ID.
// @Tags domain
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param deleteDomain body types.DeleteDomainRequest true "Domain deletion request"
// @Success 200 {object} types.Response "Domain deleted successfully"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /domain [delete]
func (c *DomainsController) DeleteDomain(w http.ResponseWriter, r *http.Request) {
	var domainRequest types.DeleteDomainRequest
	if !c.parseAndValidate(w, r, &domainRequest) {
		return
	}

	err := c.service.DeleteDomain(domainRequest.ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Domain deleted successfully", nil)
}
