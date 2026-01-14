package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/tasks"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// AddApplicationDomainRequest represents a request to add a domain to an application
type AddApplicationDomainRequest struct {
	Domain string `json:"domain"`
}

// RemoveApplicationDomainRequest represents a request to remove a domain from an application
type RemoveApplicationDomainRequest struct {
	Domain string `json:"domain"`
}

// AddApplicationDomain adds a domain to an application
func (c *DeployController) AddApplicationDomain(f fuego.ContextWithBody[AddApplicationDomainRequest]) (*types.ApplicationResponse, error) {
	applicationID := f.QueryParam("id")
	if applicationID == "" {
		return nil, fuego.HTTPError{
			Err:    types.ErrMissingID,
			Status: http.StatusBadRequest,
		}
	}

	data, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, "failed to read request body", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	if data.Domain == "" {
		return nil, fuego.HTTPError{
			Err:    types.ErrMissingDomain,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	appID, err := uuid.Parse(applicationID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	// Verify application exists and user has access
	application, err := c.service.GetApplicationById(applicationID, organizationID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusNotFound,
		}
	}

	// Check domain limit (max 5 domains per application)
	existingDomains, err := c.storage.GetApplicationDomains(appID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	// Check for duplicate domain
	for _, existingDomain := range existingDomains {
		if existingDomain.Domain == data.Domain {
			return nil, fuego.HTTPError{
				Err:    types.ErrDomainAlreadyExists,
				Status: http.StatusBadRequest,
			}
		}
	}

	if len(existingDomains) >= 5 {
		return nil, fuego.HTTPError{
			Err:    types.ErrDomainLimitReached,
			Status: http.StatusBadRequest,
		}
	}

	// Add domain
	err = c.storage.AddApplicationDomains(appID, []string{data.Domain})
	if err != nil {
		c.logger.Log(logger.Error, "failed to add domain", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	// Reload application with domains
	application, err = c.service.GetApplicationById(applicationID, organizationID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ApplicationResponse{
		Status:  "success",
		Message: "Domain added successfully",
		Data:    application,
	}, nil
}

// RemoveApplicationDomain removes a domain from an application
func (c *DeployController) RemoveApplicationDomain(f fuego.ContextWithBody[RemoveApplicationDomainRequest]) (*types.ApplicationResponse, error) {
	applicationID := f.QueryParam("id")
	if applicationID == "" {
		return nil, fuego.HTTPError{
			Err:    types.ErrMissingID,
			Status: http.StatusBadRequest,
		}
	}

	data, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Error, "failed to read request body", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	if data.Domain == "" {
		return nil, fuego.HTTPError{
			Err:    types.ErrMissingDomain,
			Status: http.StatusBadRequest,
		}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	organizationID := utils.GetOrganizationID(f.Request())
	if organizationID == uuid.Nil {
		return nil, fuego.HTTPError{
			Err:    nil,
			Status: http.StatusUnauthorized,
		}
	}

	appID, err := uuid.Parse(applicationID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusBadRequest,
		}
	}

	// Verify application exists and user has access
	_, err = c.service.GetApplicationById(applicationID, organizationID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusNotFound,
		}
	}

	// Remove domain from database
	err = c.storage.RemoveApplicationDomain(appID, data.Domain)
	if err != nil {
		c.logger.Log(logger.Error, "failed to remove domain", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	// Remove domain from Caddy proxy immediately
	client := tasks.GetCaddyClient()
	if client != nil {
		if err := client.DeleteDomain(data.Domain); err != nil {
			c.logger.Log(logger.Warning, "failed to remove domain from proxy", err.Error())
			// Don't fail the request if proxy removal fails, domain is already removed from DB
		} else {
			client.Reload()
		}
	}

	// Reload application with domains
	application, err := c.service.GetApplicationById(applicationID, organizationID)
	if err != nil {
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &types.ApplicationResponse{
		Status:  "success",
		Message: "Domain removed successfully",
		Data:    application,
	}, nil
}
