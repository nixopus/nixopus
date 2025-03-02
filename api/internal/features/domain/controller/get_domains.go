package controller

import (
	"net/http"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// @Summary Get all domains
// @Description Retrieves a list of all domains.
// @Tags domain
// @Accept json
// @Produce json
// @Success 200 {object} types.Response "Success response with domains"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /domain [get]
func (c *DomainsController) GetDomains(w http.ResponseWriter, r *http.Request) {
	domains, err := c.service.GetDomains()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Domains", domains)
}
