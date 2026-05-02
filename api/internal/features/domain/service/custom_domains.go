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
	s.logger.Log(logger.Info, "add custom domain request", fmt.Sprintf("domain=%s, org_id=%s", name, orgID))

	if err := s.validateDomainName(name); err != nil {
		return nil, nil, "", err
	}

	defaultKey, err := s.storage.GetDefaultSSHKeyByOrg(orgID)
	if err != nil {
		s.logger.Log(logger.Error, "failed to get default server for org", err.Error())
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

	if isBYOS {
		machineIP := ""
		if defaultKey.Host != nil && *defaultKey.Host != "" {
			machineIP = *defaultKey.Host
		}
		if machineIP == "" {
			return nil, nil, "", fmt.Errorf("default server has no IP configured")
		}
		targetSubdomain = machineIP

		domain := buildDomain(userID, orgID, name, verificationToken, dnsProvider, targetSubdomain)
		if err := s.storage.CreateCustomDomain(domain); err != nil {
			s.logger.Log(logger.Error, "failed to create custom domain", err.Error())
			return nil, nil, "", err
		}
		instructions := s.dns.GenerateDNSInstructionsBYOS(name, machineIP, dnsProvider)
		return domain, instructions, dnsProvider, nil
	}

	if targetSubdomain == "" {
		return nil, nil, "", fmt.Errorf("no subdomain configured for this organization")
	}

	domain := buildDomain(userID, orgID, name, verificationToken, dnsProvider, targetSubdomain)
	if err := s.storage.CreateCustomDomain(domain); err != nil {
		s.logger.Log(logger.Error, "failed to create custom domain", err.Error())
		return nil, nil, "", err
	}
	instructions := s.dns.GenerateDNSInstructions(name, targetSubdomain, dnsProvider)
	return domain, instructions, dnsProvider, nil
}

func (s *DomainsService) VerifyCustomDomain(ctx context.Context, domainID, orgID uuid.UUID) (*shared_types.Domain, error) {
	s.logger.Log(logger.Info, "verify custom domain request", fmt.Sprintf("domain_id=%s", domainID))

	domain, err := s.storage.GetCustomDomainByID(domainID, orgID)
	if err != nil {
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
		s.logger.Log(logger.Error, "DNS verification failed", verifyErr.Error())
		return nil, verifyErr
	}
	if !verified {
		return nil, types.ErrDNSNotVerified
	}

	if err := s.storage.UpdateCustomDomainVerification(domainID, "dns_verified", domain.DNSProvider); err != nil {
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
		s.logger.Log(logger.Error, "failed to enqueue domain registration", err.Error())
	}

	domain.Status = "dns_verified"
	return domain, nil
}

func (s *DomainsService) RemoveCustomDomain(ctx context.Context, domainID, orgID uuid.UUID) error {
	s.logger.Log(logger.Info, "remove custom domain request", fmt.Sprintf("domain_id=%s", domainID))

	domain, err := s.storage.GetCustomDomainByID(domainID, orgID)
	if err != nil {
		return err
	}

	if err := s.storage.UpdateCustomDomainStatus(domainID, "removing"); err != nil {
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
		s.logger.Log(logger.Error, "failed to enqueue domain removal", err.Error())
	}

	return s.storage.DeleteCustomDomain(domainID)
}

func (s *DomainsService) ListCustomDomains(ctx context.Context, orgID uuid.UUID) ([]shared_types.Domain, error) {
	return s.storage.GetCustomDomainsByOrg(orgID)
}

func (s *DomainsService) CheckDNSStatus(ctx context.Context, domainID, orgID uuid.UUID) (bool, string, error) {
	domain, err := s.storage.GetCustomDomainByID(domainID, orgID)
	if err != nil {
		return false, "", err
	}

	targetSubdomain := ""
	if domain.TargetSubdomain != nil {
		targetSubdomain = *domain.TargetSubdomain
	}

	if net.ParseIP(targetSubdomain) != nil {
		verified, _ := s.dns.VerifyARecord(domain.Name, targetSubdomain)
		if verified {
			return true, "verified", nil
		}
		propagationStatus, _ := s.dns.CheckPropagationBYOS(domain.Name, targetSubdomain)
		return false, propagationStatus, nil
	}

	verified, err := s.dns.VerifyDNSConfig(domain.Name, targetSubdomain)
	if err != nil {
		return false, "not_configured", nil
	}
	if verified {
		return true, "verified", nil
	}
	propagationStatus, _ := s.dns.CheckPropagation(domain.Name)
	return false, propagationStatus, nil
}

// validateDomainName checks format then uniqueness before creating a domain.
func (s *DomainsService) validateDomainName(name string) error {
	if err := validation.NewValidator(s.storage).ValidateName(name); err != nil {
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
