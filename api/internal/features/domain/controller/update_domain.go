package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// UpdateDomain updates an existing domain.
//
// This endpoint is accessible by the authenticated user.
//
// @Summary Update a domain
// @Description Updates an existing domain in the application.
// @Tags domain
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param domain body types.UpdateDomainRequest true "Domain update request"
// @Success 200 {object} types.Domain "Success response with updated domain"
// @Failure 400 {object} types.Response "Bad request"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /domain [put]
func (c *DomainsController) UpdateDomain(w http.ResponseWriter, r *http.Request) {
	var domainRequest types.UpdateDomainRequest

	if !c.parseAndValidate(w, r, &domainRequest) {
		return
	}

	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	updated, err := c.service.UpdateDomain(domainRequest.Name, user.ID.String(), domainRequest.ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Domain updated successfully", updated)
}
