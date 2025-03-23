package controller

import (
	"math/rand"
	"net/http"
	"time"

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
// @Router /domain/all [get]
func (c *DomainsController) GetDomains(w http.ResponseWriter, r *http.Request) {
	user := c.GetUser(w, r)

	if user == nil {
		return
	}

	if err := c.validator.AccessValidator(w, r, user); err != nil {
		c.logger.Log(logger.Error, err.Error(), err.Error())
		utils.SendErrorResponse(w, err.Error(), http.StatusForbidden)
		return
	}

	domains, err := c.service.GetDomains()
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		utils.SendErrorResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.SendJSONResponse(w, "success", "Domains", domains)
}

func (c *DomainsController) GenerateRandomSubDomain(w http.ResponseWriter, r *http.Request) {
	domains, err := c.service.GetDomains()
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

	response := struct {
		Subdomain string `json:"subdomain"`
		Domain    string `json:"domain"`
	}{
		Subdomain: subdomain,
		Domain:    randomDomain.Name,
	}

	c.logger.Log(logger.Info, "Generated random subdomain", subdomain)

	utils.SendJSONResponse(w, "success", "RandomSubdomain", response)
}
