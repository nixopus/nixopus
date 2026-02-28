package controller

import (
	"context"
	"net/http"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/caddy"
	"github.com/raghavyuva/nixopus-api/internal/features/deploy/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

// AddApplicationDomainRequest represents a request to add a domain to an application
type AddApplicationDomainRequest struct {
	Domain      string `json:"domain"`
	ServiceName string `json:"service_name,omitempty"`
	Port        *int   `json:"port,omitempty"`
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

	// Resolve compose service if service_name is provided
	var composeServiceID *uuid.UUID
	if data.ServiceName != "" {
		svc, svcErr := c.storage.GetComposeServiceByName(appID, data.ServiceName)
		if svcErr != nil {
			return nil, fuego.HTTPError{
				Err:    svcErr,
				Status: http.StatusInternalServerError,
			}
		}
		if svc != nil {
			composeServiceID = &svc.ID
		}
	}

	err = c.storage.AddApplicationDomainWithService(appID, data.Domain, composeServiceID, data.Port)
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

	orgCtx := context.WithValue(f.Request().Context(), shared_types.OrganizationIDKey, organizationID.String())
	if err := caddy.RemoveDomainsWithRetry(orgCtx, nil, &c.logger, []string{data.Domain}); err != nil {
		c.logger.Log(logger.Warning, "failed to remove domain from proxy, enqueueing for retry", err.Error())
		if enqErr := caddy.EnqueuePendingRemoval(organizationID, data.Domain); enqErr != nil {
			c.logger.Log(logger.Error, "failed to enqueue pending removal", enqErr.Error())
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

// syncApplicationDomains syncs application domains to match the desired list.
// Removes domains no longer in the list (and updates Caddy), adds new domains.
func (c *DeployController) syncApplicationDomains(appID uuid.UUID, organizationID uuid.UUID, desiredDomains []string) error {
	// Normalize desired domains: trim, filter empty
	desiredSet := make(map[string]bool)
	for _, d := range desiredDomains {
		trimmed := strings.TrimSpace(d)
		if trimmed != "" {
			desiredSet[strings.ToLower(trimmed)] = true
		}
	}

	existingDomains, err := c.storage.GetApplicationDomains(appID)
	if err != nil {
		return err
	}

	existingSet := make(map[string]string) // lowercase -> actual
	for _, d := range existingDomains {
		existingSet[strings.ToLower(d.Domain)] = d.Domain
	}

	// Remove domains that are no longer desired
	for existingLower, actualDomain := range existingSet {
		if !desiredSet[existingLower] {
			if err := c.storage.RemoveApplicationDomain(appID, actualDomain); err != nil {
				return err
			}
			orgCtx := context.WithValue(c.ctx, shared_types.OrganizationIDKey, organizationID.String())
			if err := caddy.RemoveDomainsWithRetry(orgCtx, nil, &c.logger, []string{actualDomain}); err != nil {
				c.logger.Log(logger.Warning, "failed to remove domain from proxy, enqueueing for retry", err.Error())
				if enqErr := caddy.EnqueuePendingRemoval(organizationID, actualDomain); enqErr != nil {
					c.logger.Log(logger.Error, "failed to enqueue pending removal", enqErr.Error())
				}
			}
		}
	}

	// Add domains that are new
	var toAdd []string
	for desiredLower := range desiredSet {
		if _, exists := existingSet[desiredLower]; !exists {
			// Get actual casing from desired list
			for _, d := range desiredDomains {
				if strings.ToLower(strings.TrimSpace(d)) == desiredLower {
					toAdd = append(toAdd, strings.TrimSpace(d))
					break
				}
			}
		}
	}

	if len(toAdd) > 0 {
		return c.storage.AddApplicationDomains(appID, toAdd)
	}

	return nil
}

// syncComposeApplicationDomains syncs compose-specific domains, including service linkage and port overrides.
func (c *DeployController) syncComposeApplicationDomains(appID uuid.UUID, organizationID uuid.UUID, composeDomains []types.ComposeDomain) error {
	desiredSet := make(map[string]types.ComposeDomain)
	for _, cd := range composeDomains {
		trimmed := strings.TrimSpace(cd.Domain)
		if trimmed != "" {
			desiredSet[strings.ToLower(trimmed)] = cd
		}
	}

	existingDomains, err := c.storage.GetApplicationDomains(appID)
	if err != nil {
		return err
	}

	existingSet := make(map[string]string) // lowercase -> actual domain
	for _, d := range existingDomains {
		existingSet[strings.ToLower(d.Domain)] = d.Domain
	}

	for existingLower, actualDomain := range existingSet {
		if _, wanted := desiredSet[existingLower]; !wanted {
			if err := c.storage.RemoveApplicationDomain(appID, actualDomain); err != nil {
				return err
			}
			orgCtx := context.WithValue(c.ctx, shared_types.OrganizationIDKey, organizationID.String())
			if err := caddy.RemoveDomainsWithRetry(orgCtx, nil, &c.logger, []string{actualDomain}); err != nil {
				c.logger.Log(logger.Warning, "failed to remove domain from proxy, enqueueing for retry", err.Error())
				if enqErr := caddy.EnqueuePendingRemoval(organizationID, actualDomain); enqErr != nil {
					c.logger.Log(logger.Error, "failed to enqueue pending removal", enqErr.Error())
				}
			}
		}
	}

	for desiredLower, cd := range desiredSet {
		var composeServiceID *uuid.UUID
		var port *int

		if cd.ServiceName != "" {
			svc, svcErr := c.storage.GetComposeServiceByName(appID, cd.ServiceName)
			if svcErr != nil {
				return svcErr
			}
			if svc != nil {
				composeServiceID = &svc.ID
			}
		}
		if cd.Port > 0 {
			p := cd.Port
			port = &p
		}

		if _, exists := existingSet[desiredLower]; exists {
			if err := c.storage.UpdateApplicationDomainService(appID, cd.Domain, composeServiceID, port); err != nil {
				return err
			}
		} else {
			if err := c.storage.AddApplicationDomainWithService(appID, strings.TrimSpace(cd.Domain), composeServiceID, port); err != nil {
				return err
			}
		}
	}

	return nil
}

// GetComposeServices returns the discovered compose services for an application.
func (c *DeployController) GetComposeServices(f fuego.ContextNoBody) (*shared_types.Response, error) {
	applicationID := f.QueryParam("id")
	if applicationID == "" {
		return nil, fuego.HTTPError{
			Err:    types.ErrMissingID,
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

	services, err := c.storage.GetComposeServices(appID)
	if err != nil {
		c.logger.Log(logger.Error, "failed to get compose services", err.Error())
		return nil, fuego.HTTPError{
			Err:    err,
			Status: http.StatusInternalServerError,
		}
	}

	return &shared_types.Response{
		Status:  "success",
		Message: "Compose services fetched successfully",
		Data:    services,
	}, nil
}
