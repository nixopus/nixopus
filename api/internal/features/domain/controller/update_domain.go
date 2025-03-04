package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

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
