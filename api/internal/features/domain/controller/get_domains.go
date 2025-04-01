package controller

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"

	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

func (c *DomainsController) GetDomains(f fuego.ContextNoBody) (*shared_types.Response, error) {
	organization_id := f.QueryParam("id")

	w, r := f.Response(), f.Request()

	user := utils.GetUser(w, r)

	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	c.logger.Log(logger.Info, "fetching domains", fmt.Sprintf("organization_id: %s", organization_id))

	domains, err := c.service.GetDomains(organization_id, user.ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Domains fetched successfully",
		Data:    domains,
	}, nil
}

func (c *DomainsController) GenerateRandomSubDomain(f fuego.ContextNoBody) (*shared_types.Response, error) {
	w, r := f.Response(), f.Request()

	organization_id := f.QueryParam("id")

	domains, err := c.service.GetDomains(organization_id, utils.GetUser(w, r).ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	if len(domains) == 0 {
		c.logger.Log(logger.Error, "no domains available", "")
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusBadRequest,
		}
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

	return &shared_types.Response{
		Status:  "success",
		Message: "Random subdomain generated successfully",
		Data:    response,
	}, nil
}
