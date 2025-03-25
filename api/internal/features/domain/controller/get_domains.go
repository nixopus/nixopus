package controller

import (
	"math/rand"
	"net/http"
	"time"

	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// @Summary Get all domains
// @Description Retrieves a list of all domains.
// @Tags domain
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.Response "Success response with domains"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /domains [get]
func (c *DomainsController) GetDomains(w http.ResponseWriter, r *http.Request) {
	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	organization_id := r.URL.Query().Get("id")

	if err := c.validator.AccessValidator(w, r, user, nil); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusForbidden)
		return
	}

	domains, err := c.service.GetDomains(organization_id, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Domains", domains)
}

// @Summary Generate a random subdomain
// @Description Generates a random subdomain by taking a random domain from the list of all domains and appending a random string to it.
// @Tags domain
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} types.RandomSubdomainResponse "Success response with random subdomain"
// @Failure 404 {object} types.Response "No domains available"
// @Failure 500 {object} types.Response "Internal server error"
// @Router /domain/generate [get]
func (c *DomainsController) GenerateRandomSubDomain(w http.ResponseWriter, r *http.Request) {
	organization_id := r.URL.Query().Get("id")
	domains, err := c.service.GetDomains(organization_id, c.GetUser(w, r).ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(domains) == 0 {
		c.logger.Log(logger.Error, "no domains available", "")
		utils.SendErrorResponse(w, "no domains available", http.StatusNotFound)
		return
	}

	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	const prefixLength = 8
	randomPrefix := make([]byte, prefixLength)

	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)

	for i := range randomPrefix {
		randomPrefix[i] = charset[random.Intn(len(charset))]
	}

	randomDomain := domains[random.Intn(len(domains))]

	subdomain := string(randomPrefix) + "." + randomDomain.Name

	response := types.RandomSubdomainResponse{
		Subdomain: subdomain,
		Domain:    randomDomain.Name,
	}

	c.logger.Log(logger.Info, "Generated random subdomain", subdomain)

	utils.SendJSONResponse(w, "success", "RandomSubdomain", response)
}
