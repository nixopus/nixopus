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

	validator := validation.NewValidator(s.storage)
	if err := validator.ValidateName(name); err != nil {
		return nil, nil, "", err
	}

	existing, err := s.storage.GetCustomDomainByName(name)
	if err != nil {
		return nil, nil, "", err
	}
	if existing != nil {
		return nil, nil, "", types.ErrDomainAlreadyExists
	}

	var provisionDetails shared_types.UserProvisionDetails
	q := s.store.DB.NewSelect().Model(&provisionDetails)
	serverIDStr, _ := ctx.Value(shared_types.ServerIDKey).(string)
	if serverIDStr != "" {
		if sid, parseErr := uuid.Parse(serverIDStr); parseErr == nil {
			q = q.Where("ssh_key_id = ? AND organization_id = ?", sid, orgID)
		} else {
			q = q.Where("organization_id = ?", orgID)
		}
	} else {
		q = q.Where("organization_id = ?", orgID)
	}
	err = q.Limit(1).Scan(ctx)
	if err != nil {
		s.logger.Log(logger.Error, "failed to get provision details", err.Error())
		return nil, nil, "", fmt.Errorf("provision details not found for organization")
	}

	targetSubdomain := ""
	if provisionDetails.Subdomain != nil {
		targetSubdomain = *provisionDetails.Subdomain
	}

	dnsProvider, _ := DetectDNSProvider(name)
	verificationToken := GenerateVerificationToken()

	// BYOS path: user-owned machine — use machine IP as target instead of nixopus subdomain
	if provisionDetails.Type == "user_owned" {
		machineIP, ipErr := s.resolveMachineIP(ctx, provisionDetails)
		if ipErr != nil {
			s.logger.Log(logger.Error, "failed to resolve machine IP for BYOS domain", ipErr.Error())
			return nil, nil, "", fmt.Errorf("machine IP not configured: please ensure your server IP is set")
		}
		targetSubdomain = machineIP

		domain := &shared_types.Domain{
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
		if err := s.storage.CreateCustomDomain(domain); err != nil {
			s.logger.Log(logger.Error, "failed to create custom domain", err.Error())
			return nil, nil, "", err
		}
		instructions := GenerateDNSInstructionsBYOS(name, machineIP, dnsProvider)
		return domain, instructions, dnsProvider, nil
	}

	// Managed path — require nixopus-assigned subdomain
	if targetSubdomain == "" {
		return nil, nil, "", fmt.Errorf("no subdomain configured for this organization")
	}

	domain := &shared_types.Domain{
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

	if err := s.storage.CreateCustomDomain(domain); err != nil {
		s.logger.Log(logger.Error, "failed to create custom domain", err.Error())
		return nil, nil, "", err
	}

	instructions := GenerateDNSInstructions(name, targetSubdomain, dnsProvider)
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

	// BYOS path: target_subdomain holds a machine IP (not a nixopus subdomain)
	if net.ParseIP(targetSubdomain) != nil {
		verified, verifyErr = VerifyDNSRecordMatchesMachineIP(domain.Name, targetSubdomain)
	} else {
		verified, verifyErr = VerifyDNSConfiguration(domain.Name, targetSubdomain)
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

	// Enqueue (no-op worker — domain becomes available for app assignment)
	payload := queue.CustomDomainPayload{
		DomainID:  domainID.String(),
		Domain:    domain.Name,
		Subdomain: targetSubdomain,
	}

	// For managed machines, also populate ServerID and GuestIP for future worker use
	if net.ParseIP(targetSubdomain) == nil && targetSubdomain != "" {
		var provisionDetails shared_types.UserProvisionDetails
		if dbErr := s.store.DB.NewSelect().
			Model(&provisionDetails).
			Where("subdomain = ?", targetSubdomain).
			Limit(1).
			Scan(ctx); dbErr == nil {
			if provisionDetails.ServerID != nil {
				payload.ServerID = provisionDetails.ServerID.String()
			}
			if provisionDetails.GuestIP != nil {
				payload.GuestIP = *provisionDetails.GuestIP
			}
		}
	}

	err = queue.EnqueueRegisterCustomDomain(ctx, payload)
	if err != nil {
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

	// Look up server_id for per-server queue routing.
	if domain.TargetSubdomain != nil && *domain.TargetSubdomain != "" {
		var provisionDetails shared_types.UserProvisionDetails
		if dbErr := s.store.DB.NewSelect().
			Model(&provisionDetails).
			Where("subdomain = ?", *domain.TargetSubdomain).
			Limit(1).
			Scan(ctx); dbErr == nil && provisionDetails.ServerID != nil {
			removePayload.ServerID = provisionDetails.ServerID.String()
		}
	}

	err = queue.EnqueueRemoveCustomDomain(ctx, removePayload)
	if err != nil {
		s.logger.Log(logger.Error, "failed to enqueue domain removal", err.Error())
	}

	return s.storage.DeleteCustomDomain(domainID)
}

func (s *DomainsService) ListCustomDomains(ctx context.Context, orgID uuid.UUID) ([]shared_types.Domain, error) {
	return s.storage.GetCustomDomainsByOrg(orgID)
}

// resolveMachineIP returns the public IP for a BYOS machine.
// Prefers GuestIP on the provision row; falls back to ssh_keys.host.
func (s *DomainsService) resolveMachineIP(ctx context.Context, provision shared_types.UserProvisionDetails) (string, error) {
	if provision.GuestIP != nil && *provision.GuestIP != "" {
		return *provision.GuestIP, nil
	}
	if provision.SSHKeyID != nil {
		var key shared_types.SSHKey
		err := s.store.DB.NewSelect().
			Model(&key).
			Column("host").
			Where("id = ?", *provision.SSHKeyID).
			Scan(ctx)
		if err == nil && key.Host != nil && *key.Host != "" {
			return *key.Host, nil
		}
	}
	return "", fmt.Errorf("could not resolve machine IP: GuestIP and ssh_keys.host are both empty")
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

	// BYOS path: target_subdomain holds machine IP
	if net.ParseIP(targetSubdomain) != nil {
		verified, _ := VerifyDNSRecordMatchesMachineIP(domain.Name, targetSubdomain)
		if verified {
			return true, "verified", nil
		}
		propagationStatus, _ := CheckDNSPropagationBYOS(domain.Name, targetSubdomain)
		return false, propagationStatus, nil
	}

	// Managed path — unchanged
	verified, err := VerifyDNSConfiguration(domain.Name, targetSubdomain)
	if err != nil {
		return false, "not_configured", nil
	}
	if verified {
		return true, "verified", nil
	}
	propagationStatus, _ := CheckDNSPropagation(domain.Name)
	return false, propagationStatus, nil
}
