package controller

import (
	"fmt"
	"net/http"

	"github.com/go-fuego/fuego"
	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/domain/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/nixopus/nixopus/api/internal/utils"
)

func (c *DomainsController) HandleAddCustomDomain(f fuego.ContextWithBody[types.AddCustomDomainRequest]) (*types.DNSSetupResponse, error) {
	req, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("add custom domain: invalid request body: %v", err), "")
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "add custom domain: authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		c.logger.Log(logger.Error, "add custom domain: organization ID is required", user.ID.String())
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	ctx := fmt.Sprintf("user_id=%s org_id=%s name=%s", user.ID, orgID, req.Name)
	c.logger.Log(logger.Info, "add custom domain: submitting", ctx)

	domain, instructions, dnsProvider, err := c.service.AddCustomDomain(f.Context(), user.ID, orgID, req.Name)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("add custom domain: %v", err), ctx)
		return nil, mapCustomDomainError(err)
	}

	c.logger.Log(logger.Info, "add custom domain: success", fmt.Sprintf("%s domain_id=%s", ctx, domain.ID))

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
		c.logger.Log(logger.Error, "list custom domains: authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		c.logger.Log(logger.Error, "list custom domains: organization ID is required", user.ID.String())
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	ctx := fmt.Sprintf("user_id=%s org_id=%s", user.ID, orgID)
	c.logger.Log(logger.Info, "list custom domains: fetching", ctx)

	domains, err := c.service.ListCustomDomains(f.Context(), orgID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("list custom domains: %v", err), ctx)
		return nil, fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}

	c.logger.Log(logger.Info, "list custom domains: success", fmt.Sprintf("%s count=%d", ctx, len(domains)))

	return &types.CustomDomainListResponse{
		Status:  "success",
		Message: "Custom domains retrieved successfully",
		Data:    domains,
	}, nil
}

func (c *DomainsController) HandleVerifyCustomDomain(f fuego.ContextWithBody[types.VerifyCustomDomainRequest]) (*types.CustomDomainResponse, error) {
	req, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("verify custom domain: invalid request body: %v", err), "")
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "verify custom domain: authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		c.logger.Log(logger.Error, "verify custom domain: organization ID is required", user.ID.String())
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	domainID, err := uuid.Parse(req.ID)
	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("verify custom domain: invalid domain id: %v", err), user.ID.String())
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	ctx := fmt.Sprintf("user_id=%s org_id=%s domain_id=%s", user.ID, orgID, domainID)
	c.logger.Log(logger.Info, "verify custom domain: submitting", ctx)

	domain, err := c.service.VerifyCustomDomain(f.Context(), domainID, orgID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("verify custom domain: %v", err), ctx)
		return nil, mapCustomDomainError(err)
	}

	c.logger.Log(logger.Info, "verify custom domain: success", ctx)

	return &types.CustomDomainResponse{
		Status:  "success",
		Message: "Domain DNS verified successfully",
		Data:    domain,
	}, nil
}

func (c *DomainsController) HandleRemoveCustomDomain(f fuego.ContextWithBody[types.RemoveCustomDomainRequest]) (*types.MessageResponse, error) {
	req, err := f.Body()
	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("remove custom domain: invalid request body: %v", err), "")
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "remove custom domain: authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		c.logger.Log(logger.Error, "remove custom domain: organization ID is required", user.ID.String())
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	domainID, err := uuid.Parse(req.ID)
	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("remove custom domain: invalid domain id: %v", err), user.ID.String())
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	ctx := fmt.Sprintf("user_id=%s org_id=%s domain_id=%s", user.ID, orgID, domainID)
	c.logger.Log(logger.Info, "remove custom domain: submitting", ctx)

	if err := c.service.RemoveCustomDomain(f.Context(), domainID, orgID); err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("remove custom domain: %v", err), ctx)
		return nil, mapCustomDomainError(err)
	}

	c.logger.Log(logger.Info, "remove custom domain: success", ctx)

	return &types.MessageResponse{
		Status:  "success",
		Message: "Custom domain removed successfully",
	}, nil
}

func (c *DomainsController) HandleCheckDNSStatus(f fuego.ContextNoBody) (*types.DNSCheckResponse, error) {
	user := utils.GetUser(f.Response(), f.Request())
	if user == nil {
		c.logger.Log(logger.Error, "check DNS status: authentication required", "")
		return nil, fuego.UnauthorizedError{Detail: "authentication required"}
	}

	orgID := utils.GetOrganizationID(f.Request())
	if orgID == uuid.Nil {
		c.logger.Log(logger.Error, "check DNS status: organization ID is required", user.ID.String())
		return nil, fuego.BadRequestError{Detail: "organization ID is required"}
	}

	domainIDStr := f.QueryParam("id")
	if domainIDStr == "" {
		c.logger.Log(logger.Debug, "check DNS status: missing id query parameter", user.ID.String())
		return nil, fuego.BadRequestError{
			Detail: shared_types.ErrFailedToGetUserFromContext.Error(),
			Err:    shared_types.ErrFailedToGetUserFromContext,
		}
	}

	domainID, err := uuid.Parse(domainIDStr)
	if err != nil {
		c.logger.Log(logger.Debug, fmt.Sprintf("check DNS status: invalid domain id: %v", err), user.ID.String())
		return nil, fuego.BadRequestError{Detail: err.Error(), Err: err}
	}

	ctx := fmt.Sprintf("user_id=%s org_id=%s domain_id=%s", user.ID, orgID, domainID)
	c.logger.Log(logger.Debug, "check DNS status: checking", ctx)

	verified, dnsStatus, err := c.service.CheckDNSStatus(f.Context(), domainID, orgID)
	if err != nil {
		c.logger.Log(logger.Error, fmt.Sprintf("check DNS status: %v", err), ctx)
		return nil, mapCustomDomainError(err)
	}

	c.logger.Log(logger.Debug, fmt.Sprintf("check DNS status: done verified=%v dns_status=%s", verified, dnsStatus), ctx)

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

func mapCustomDomainError(err error) error {
	switch err {
	case types.ErrDomainAlreadyExists:
		return fuego.ConflictError{Detail: err.Error(), Err: err}
	case types.ErrCustomDomainNotFound:
		return fuego.NotFoundError{Detail: err.Error(), Err: err}
	case types.ErrDNSNotVerified:
		return fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusPreconditionFailed}
	case types.ErrSubscriptionRequired:
		return fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusPaymentRequired}
	case types.ErrMaxCustomDomainsReached:
		return fuego.ForbiddenError{Detail: err.Error(), Err: err}
	case types.ErrInvalidCustomDomain:
		return fuego.BadRequestError{Detail: err.Error(), Err: err}
	default:
		if isInvalidDomainError(err) {
			return fuego.BadRequestError{Detail: err.Error(), Err: err}
		}
		return fuego.HTTPError{Err: err, Detail: err.Error(), Status: http.StatusInternalServerError}
	}
}
