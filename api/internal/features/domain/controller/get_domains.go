package controller

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/domain/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *DomainsController) GetDomains(f fuego.ContextNoBody) (*types.ListDomainsResponse, error) {
	w, r := f.Response(), f.Request()

	organization_id := utils.GetOrganizationID(r)
	if organization_id == uuid.Nil {
		c.logger.Log(logger.Error, "invalid organization id", "")
		return nil, fuego.BadRequestError{
			Detail: types.ErrMissingID.Error(),
			Err:    types.ErrMissingID,
		}
	}

	user := utils.GetUser(w, r)
	if user == nil {
		c.logger.Log(logger.Error, "unauthorized user", "")
		return nil, fuego.UnauthorizedError{
			Detail: types.ErrAccessDenied.Error(),
			Err:    types.ErrAccessDenied,
		}
	}

	domainType := f.QueryParam("type")

	c.logger.Log(logger.Info, "fetching domains", fmt.Sprintf("organization_id: %s", organization_id))

	domains, err := c.service.GetDomains(organization_id.String(), user.ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")

		if isPermissionError(err) {
			return nil, fuego.ForbiddenError{
				Detail: err.Error(),
				Err:    err,
			}
		}

		if err == types.ErrDomainNotFound {
			return nil, fuego.NotFoundError{
				Detail: err.Error(),
				Err:    err,
			}
		}

		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	if domainType != "" {
		filtered := make([]shared_types.Domain, 0)
		for _, d := range domains {
			if d.Type == domainType {
				filtered = append(filtered, d)
			}
		}
		domains = filtered
	}

	return &types.ListDomainsResponse{
		Status:  "success",
		Message: "Domains fetched successfully",
		Data:    domains,
	}, nil
}

func (c *DomainsController) GenerateRandomSubDomain(f fuego.ContextNoBody) (*types.RandomSubdomainResponseWrapper, error) {
	w, r := f.Response(), f.Request()

	organization_id := utils.GetOrganizationID(r)
	if organization_id == uuid.Nil {
		c.logger.Log(logger.Error, "invalid organization id", "")
		return nil, fuego.BadRequestError{
			Detail: types.ErrMissingID.Error(),
			Err:    types.ErrMissingID,
		}
	}

	user := utils.GetUser(w, r)
	if user == nil {
		c.logger.Log(logger.Error, "unauthorized user", "")
		return nil, fuego.UnauthorizedError{
			Detail: types.ErrAccessDenied.Error(),
			Err:    types.ErrAccessDenied,
		}
	}

	domains, err := c.service.GetDomains(organization_id.String(), user.ID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")

		if isPermissionError(err) {
			return nil, fuego.ForbiddenError{
				Detail: err.Error(),
				Err:    err,
			}
		}

		if err == types.ErrDomainNotFound {
			return nil, fuego.NotFoundError{
				Detail: err.Error(),
				Err:    err,
			}
		}

		return nil, fuego.HTTPError{
			Err:    err,
			Detail: err.Error(),
			Status: http.StatusInternalServerError,
		}
	}

	if len(domains) == 0 {
		c.logger.Log(logger.Error, "no domains available for subdomain generation", "")
		return nil, fuego.NotFoundError{
			Detail: types.ErrDomainNotFound.Error(),
			Err:    types.ErrDomainNotFound,
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

	return &types.RandomSubdomainResponseWrapper{
		Status:  "success",
		Message: "Random subdomain generated successfully",
		Data:    response,
	}, nil
}
