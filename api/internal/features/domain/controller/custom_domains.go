package controller

import (
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
	"github.com/raghavyuva/nixopus-api/internal/utils"
)

func (c *DomainsController) HandleAddCustomDomain(f fuego.ContextWithBody[types.AddCustomDomainRequest]) (*types.DNSSetupResponse, error) {
	req, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusUnauthorized}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusBadRequest}
	}

	domain, instructions, dnsProvider, err := c.service.AddCustomDomain(f.Context(), user.ID, orgID, req.Name)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, mapCustomDomainError(err)
	}

	return &types.DNSSetupResponse{
		Status:       "success",
		Message:      "Custom domain added. Configure DNS records to complete setup.",
		Data:         domain,
		Instructions: instructions,
		DNSProvider:  dnsProvider,
	}, nil
}

func (c *DomainsController) HandleListCustomDomains(f fuego.ContextNoBody) (*types.CustomDomainListResponse, error) {
	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusUnauthorized}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusBadRequest}
	}

	domains, err := c.service.ListCustomDomains(f.Context(), orgID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, fuego.HTTPError{Err: err, Status: http.StatusInternalServerError}
	}

	return &types.CustomDomainListResponse{
		Status:  "success",
		Message: "Custom domains retrieved successfully",
		Data:    domains,
	}, nil
}

func (c *DomainsController) HandleVerifyCustomDomain(f fuego.ContextWithBody[types.VerifyCustomDomainRequest]) (*types.CustomDomainResponse, error) {
	req, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusUnauthorized}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusBadRequest}
	}

	domainID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
	}

	domain, err := c.service.VerifyCustomDomain(f.Context(), domainID, orgID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, mapCustomDomainError(err)
	}

	return &types.CustomDomainResponse{
		Status:  "success",
		Message: "Domain DNS verified successfully",
		Data:    domain,
	}, nil
}

func (c *DomainsController) HandleRemoveCustomDomain(f fuego.ContextWithBody[types.RemoveCustomDomainRequest]) (*types.MessageResponse, error) {
	req, err := f.Body()
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusUnauthorized}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusBadRequest}
	}

	domainID, err := uuid.Parse(req.ID)
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
	}

	if err := c.service.RemoveCustomDomain(f.Context(), domainID, orgID); err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, mapCustomDomainError(err)
	}

	return &types.MessageResponse{
		Status:  "success",
		Message: "Custom domain removed successfully",
	}, nil
}

func (c *DomainsController) HandleCheckDNSStatus(f fuego.ContextNoBody) (*types.DNSCheckResponse, error) {
	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusUnauthorized}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		return nil, fuego.HTTPError{Err: nil, Status: http.StatusBadRequest}
	}

	domainIDStr := f.QueryParam("id")
	if domainIDStr == "" {
		return nil, fuego.HTTPError{
			Err:    shared_types.ErrFailedToGetUserFromContext,
			Status: http.StatusBadRequest,
		}
	}

	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
	}

	verified, dnsStatus, err := c.service.CheckDNSStatus(f.Context(), domainID, orgID)
	if err != nil {
		c.logger.Log(logger.Error, err.Error(), "")
		return nil, mapCustomDomainError(err)
	}

	message := "DNS is not yet configured"
	if verified {
		message = "DNS is properly configured"
	}

	return &types.DNSCheckResponse{
		Status:    "success",
		Message:   message,
		Verified:  verified,
		DNSStatus: dnsStatus,
	}, nil
}

func mapCustomDomainError(err error) fuego.HTTPError {
	switch err {
	case types.ErrDomainAlreadyExists:
		return fuego.HTTPError{Err: err, Status: http.StatusConflict}
	case types.ErrCustomDomainNotFound:
		return fuego.HTTPError{Err: err, Status: http.StatusNotFound}
	case types.ErrDNSNotVerified:
		return fuego.HTTPError{Err: err, Status: http.StatusPreconditionFailed}
	case types.ErrSubscriptionRequired:
		return fuego.HTTPError{Err: err, Status: http.StatusPaymentRequired}
	case types.ErrMaxCustomDomainsReached:
		return fuego.HTTPError{Err: err, Status: http.StatusForbidden}
	case types.ErrInvalidCustomDomain:
		return fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
	default:
		if isInvalidDomainError(err) {
			return fuego.HTTPError{Err: err, Status: http.StatusBadRequest}
		}
		return fuego.HTTPError{Err: err, Status: http.StatusInternalServerError}
	}
}
