package service

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/domain/types"
	"github.com/nixopus/nixopus/api/internal/features/domain/validation"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/queue"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (s *DomainsService) AddCustomDomain(ctx context.Context, userID, orgID uuid.UUID, name string) (*shared_types.Domain, []types.DNSInstruction, string, error) {
	ctxStr := fmt.Sprintf("user_id=%s org_id=%s name=%s", userID, orgID, name)
	s.logger.Log(logger.Info, "add custom domain: start", ctxStr)

	if err := s.validateDomainName(name); err != nil {
		if err == types.ErrDomainAlreadyExists {
			s.logger.Log(logger.Debug, "add custom domain: domain already exists", ctxStr)
		} else {
			s.logger.Log(logger.Error, fmt.Sprintf("add custom domain: validate name: %v", err), ctxStr)
		}
		return nil, nil, "", err
	}

	defaultKey, err := s.storage.GetDefaultSSHKeyByOrg(orgID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("add custom domain: get default SSH key for org: %v", err), ctxStr)
		return nil, nil, "", fmt.Errorf("no default server configured for organization")
	}

	provisionDetails, provErr := s.storage.GetProvisionDetailsBySSHKeyAndOrg(defaultKey.ID, orgID)

	isBYOS := provErr == nil && provisionDetails != nil && provisionDetails.Type == "user_owned"

	targetSubdomain := ""
	if !isBYOS && provErr == nil && provisionDetails != nil && provisionDetails.Subdomain != nil {
		targetSubdomain = *provisionDetails.Subdomain
	}

	dnsProvider, _ := s.dns.DetectProvider(name)
	verificationToken := s.dns.GenerateToken()
	s.logger.Log(logger.Debug, fmt.Sprintf("add custom domain: detected dns_provider=%s is_byos=%v", dnsProvider, isBYOS), ctxStr)

	if isBYOS {
		machineIP := ""
		if defaultKey.Host != nil && *defaultKey.Host != "" {
			machineIP = *defaultKey.Host
		}
		if machineIP == "" {
			s.logger.Log(logger.Error, "add custom domain: BYOS default server has no IP", ctxStr)
			return nil, nil, "", fmt.Errorf("default server has no IP configured")
		}
		targetSubdomain = machineIP

		domain := buildDomain(userID, orgID, name, verificationToken, dnsProvider, targetSubdomain)
		if err := s.storage.CreateCustomDomain(domain); err != nil {
			s.logger.Log(logger.Error, fmt.Sprintf("add custom domain: create (BYOS): %v", err), ctxStr)
			return nil, nil, "", err
		}
		instructions := s.dns.GenerateDNSInstructionsBYOS(name, machineIP, dnsProvider)
		s.logger.Log(logger.Info, "add custom domain: success (BYOS)", fmt.Sprintf("%s domain_id=%s", ctxStr, domain.ID))
		return domain, instructions, dnsProvider, nil
	}

	if targetSubdomain == "" {
		s.logger.Log(logger.Error, "add custom domain: no subdomain configured for organization", ctxStr)
		return nil, nil, "", fmt.Errorf("no subdomain configured for this organization")
	}

	domain := buildDomain(userID, orgID, name, verificationToken, dnsProvider, targetSubdomain)
	if err := s.storage.CreateCustomDomain(domain); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("add custom domain: create: %v", err), ctxStr)
		return nil, nil, "", err
	}
	instructions := s.dns.GenerateDNSInstructions(name, targetSubdomain, dnsProvider)
	s.logger.Log(logger.Info, "add custom domain: success", fmt.Sprintf("%s domain_id=%s target_subdomain=%s", ctxStr, domain.ID, targetSubdomain))
	return domain, instructions, dnsProvider, nil
}

func (s *DomainsService) VerifyCustomDomain(ctx context.Context, domainID, orgID uuid.UUID) (*shared_types.Domain, error) {
	ctxStr := fmt.Sprintf("org_id=%s domain_id=%s", orgID, domainID)
	s.logger.Log(logger.Info, "verify custom domain: start", ctxStr)

	domain, err := s.storage.GetCustomDomainByID(domainID, orgID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("verify custom domain: load domain: %v", err), ctxStr)
		return nil, err
	}

	targetSubdomain := ""
	if domain.TargetSubdomain != nil {
		targetSubdomain = *domain.TargetSubdomain
	}

	var verified bool
	var verifyErr error

	if net.ParseIP(targetSubdomain) != nil {
		verified, verifyErr = s.dns.VerifyARecord(domain.Name, targetSubdomain)
	} else {
		verified, verifyErr = s.dns.VerifyDNSConfig(domain.Name, targetSubdomain)
	}

	if verifyErr != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("verify custom domain: DNS check error: %v", verifyErr), fmt.Sprintf("%s name=%s", ctxStr, domain.Name))
		return nil, verifyErr
	}
	if !verified {
		s.logger.Log(logger.Debug, "verify custom domain: DNS not verified yet", fmt.Sprintf("%s name=%s target=%s", ctxStr, domain.Name, targetSubdomain))
		return nil, types.ErrDNSNotVerified
	}

	if err := s.storage.UpdateCustomDomainVerification(domainID, "dns_verified", domain.DNSProvider); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("verify custom domain: update verification status: %v", err), ctxStr)
		return nil, err
	}

	payload := queue.CustomDomainPayload{
		DomainID:  domainID.String(),
		Domain:    domain.Name,
		Subdomain: targetSubdomain,
	}

	if net.ParseIP(targetSubdomain) == nil && targetSubdomain != "" {
		if provisionDetails, dbErr := s.storage.GetProvisionDetailsBySubdomain(targetSubdomain); dbErr == nil && provisionDetails != nil {
			if provisionDetails.ServerID != nil {
				payload.ServerID = provisionDetails.ServerID.String()
			}
			if provisionDetails.GuestIP != nil {
				payload.GuestIP = *provisionDetails.GuestIP
			}
		}
	}

	if err := s.queue.EnqueueRegisterCustomDomain(ctx, payload); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("verify custom domain: enqueue registration: %v", err), ctxStr)
	}

	domain.Status = "dns_verified"
	s.logger.Log(logger.Info, "verify custom domain: success", fmt.Sprintf("%s name=%s", ctxStr, domain.Name))
	return domain, nil
}

func (s *DomainsService) RemoveCustomDomain(ctx context.Context, domainID, orgID uuid.UUID) error {
	ctxStr := fmt.Sprintf("org_id=%s domain_id=%s", orgID, domainID)
	s.logger.Log(logger.Info, "remove custom domain: start", ctxStr)

	domain, err := s.storage.GetCustomDomainByID(domainID, orgID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("remove custom domain: load domain: %v", err), ctxStr)
		return err
	}

	if err := s.storage.UpdateCustomDomainStatus(domainID, "removing"); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("remove custom domain: set status removing: %v", err), ctxStr)
		return err
	}

	removePayload := queue.RemoveCustomDomainPayload{
		DomainID: domainID.String(),
		Domain:   domain.Name,
	}

	if domain.TargetSubdomain != nil && *domain.TargetSubdomain != "" {
		if provisionDetails, dbErr := s.storage.GetProvisionDetailsBySubdomain(*domain.TargetSubdomain); dbErr == nil && provisionDetails != nil && provisionDetails.ServerID != nil {
			removePayload.ServerID = provisionDetails.ServerID.String()
		}
	}

	if err := s.queue.EnqueueRemoveCustomDomain(ctx, removePayload); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("remove custom domain: enqueue removal: %v", err), ctxStr)
	}

	if err := s.storage.DeleteCustomDomain(domainID); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("remove custom domain: delete: %v", err), ctxStr)
		return err
	}
	s.logger.Log(logger.Info, "remove custom domain: success", fmt.Sprintf("%s name=%s", ctxStr, domain.Name))
	return nil
}

func (s *DomainsService) ListCustomDomains(ctx context.Context, orgID uuid.UUID) ([]shared_types.Domain, error) {
	ctxStr := fmt.Sprintf("org_id=%s", orgID)
	s.logger.Log(logger.Debug, "list custom domains: storage lookup", ctxStr)

	domains, err := s.storage.GetCustomDomainsByOrg(orgID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("list custom domains: %v", err), ctxStr)
		return nil, err
	}
	s.logger.Log(logger.Info, "list custom domains: success", fmt.Sprintf("%s count=%d", ctxStr, len(domains)))
	return domains, nil
}

func (s *DomainsService) CheckDNSStatus(ctx context.Context, domainID, orgID uuid.UUID) (bool, string, error) {
	ctxStr := fmt.Sprintf("org_id=%s domain_id=%s", orgID, domainID)
	s.logger.Log(logger.Debug, "check DNS status: start", ctxStr)

	domain, err := s.storage.GetCustomDomainByID(domainID, orgID)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("check DNS status: load domain: %v", err), ctxStr)
		return false, "", err
	}

	targetSubdomain := ""
	if domain.TargetSubdomain != nil {
		targetSubdomain = *domain.TargetSubdomain
	}

	if net.ParseIP(targetSubdomain) != nil {
		verified, _ := s.dns.VerifyARecord(domain.Name, targetSubdomain)
		if verified {
			s.logger.Log(logger.Debug, "check DNS status: BYOS A record verified", fmt.Sprintf("%s name=%s", ctxStr, domain.Name))
			return true, "verified", nil
		}
		propagationStatus, _ := s.dns.CheckPropagationBYOS(domain.Name, targetSubdomain)
		s.logger.Log(logger.Debug, fmt.Sprintf("check DNS status: BYOS result status=%s", propagationStatus), fmt.Sprintf("%s name=%s", ctxStr, domain.Name))
		return false, propagationStatus, nil
	}

	verified, err := s.dns.VerifyDNSConfig(domain.Name, targetSubdomain)
	if err != nil {
		s.logger.Log(logger.Debug, fmt.Sprintf("check DNS status: verify config error (treating as not configured): %v", err), fmt.Sprintf("%s name=%s", ctxStr, domain.Name))
		return false, "not_configured", nil
	}
	if verified {
		s.logger.Log(logger.Debug, "check DNS status: DNS config verified", fmt.Sprintf("%s name=%s", ctxStr, domain.Name))
		return true, "verified", nil
	}
	propagationStatus, _ := s.dns.CheckPropagation(domain.Name)
	s.logger.Log(logger.Debug, fmt.Sprintf("check DNS status: propagation status=%s", propagationStatus), fmt.Sprintf("%s name=%s", ctxStr, domain.Name))
	return false, propagationStatus, nil
}

// validateDomainName checks format then uniqueness before creating a domain.
func (s *DomainsService) validateDomainName(name string) error {
	if err := validation.NewValidatorWithLogger(s.storage, &s.logger).ValidateName(name); err != nil {
		return err
	}
	existing, err := s.storage.GetCustomDomainByName(name)
	if err != nil {
		return err
	}
	if existing != nil {
		return types.ErrDomainAlreadyExists
	}
	return nil
}

// buildDomain constructs a new Domain value with common fields populated.
func buildDomain(userID, orgID uuid.UUID, name, verificationToken, dnsProvider, targetSubdomain string) *shared_types.Domain {
	return &shared_types.Domain{
		ID:                uuid.New(),
		UserID:            userID,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
		Name:              name,
		OrganizationID:    orgID,
		Type:              "custom",
		Status:            "pending_dns",
		VerificationToken: &verificationToken,
		DNSProvider:       &dnsProvider,
		TargetSubdomain:   &targetSubdomain,
	}
}
