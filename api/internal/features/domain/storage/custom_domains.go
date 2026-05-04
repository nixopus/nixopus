package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/domain/types"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

func (s *DomainStorage) GetDefaultSSHKeyByOrg(orgID uuid.UUID) (*shared_types.SSHKey, error) {
	ctxStr := fmt.Sprintf("org_id=%s", orgID)
	s.storageLog(logger.Debug, "storage: GetDefaultSSHKeyByOrg", ctxStr)

	var key shared_types.SSHKey
	err := s.getDB().NewSelect().
		Model(&key).
		Where("organization_id = ?", orgID).
		Where("is_default = ?", true).
		Where("is_active = ?", true).
		Where("deleted_at IS NULL").
		Limit(1).
		Scan(s.Ctx)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: GetDefaultSSHKeyByOrg: %v", err), ctxStr)
		return nil, err
	}
	return &key, nil
}

func (s *DomainStorage) GetProvisionDetailsBySSHKeyAndOrg(sshKeyID, orgID uuid.UUID) (*shared_types.UserProvisionDetails, error) {
	ctxStr := fmt.Sprintf("ssh_key_id=%s org_id=%s", sshKeyID, orgID)
	s.storageLog(logger.Debug, "storage: GetProvisionDetailsBySSHKeyAndOrg", ctxStr)

	var details shared_types.UserProvisionDetails
	err := s.getDB().NewSelect().
		Model(&details).
		Where("ssh_key_id = ?", sshKeyID).
		Where("organization_id = ?", orgID).
		Limit(1).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.storageLog(logger.Debug, "storage: GetProvisionDetailsBySSHKeyAndOrg not found", ctxStr)
			return nil, nil
		}
		s.storageLog(logger.Error, fmt.Sprintf("storage: GetProvisionDetailsBySSHKeyAndOrg: %v", err), ctxStr)
		return nil, err
	}
	return &details, nil
}

func (s *DomainStorage) GetProvisionDetailsBySubdomain(subdomain string) (*shared_types.UserProvisionDetails, error) {
	ctxStr := fmt.Sprintf("subdomain=%s", subdomain)
	s.storageLog(logger.Debug, "storage: GetProvisionDetailsBySubdomain", ctxStr)

	var details shared_types.UserProvisionDetails
	err := s.getDB().NewSelect().
		Model(&details).
		Where("subdomain = ?", subdomain).
		Limit(1).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.storageLog(logger.Debug, "storage: GetProvisionDetailsBySubdomain not found", ctxStr)
			return nil, nil
		}
		s.storageLog(logger.Error, fmt.Sprintf("storage: GetProvisionDetailsBySubdomain: %v", err), ctxStr)
		return nil, err
	}
	return &details, nil
}

func (s *DomainStorage) CreateCustomDomain(domain *shared_types.Domain) error {
	ctxStr := fmt.Sprintf("org_id=%s name=%s", domain.OrganizationID, domain.Name)
	s.storageLog(logger.Debug, "storage: CreateCustomDomain", ctxStr)

	domain.Type = "custom"
	_, err := s.getDB().NewInsert().Model(domain).Exec(s.Ctx)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: CreateCustomDomain: %v", err), fmt.Sprintf("%s id=%s", ctxStr, domain.ID))
		return err
	}
	s.storageLog(logger.Info, "storage: CreateCustomDomain ok", fmt.Sprintf("id=%s %s", domain.ID, ctxStr))
	return nil
}

func (s *DomainStorage) GetCustomDomainsByOrg(orgID uuid.UUID) ([]shared_types.Domain, error) {
	ctxStr := fmt.Sprintf("org_id=%s", orgID)
	s.storageLog(logger.Debug, "storage: GetCustomDomainsByOrg", ctxStr)

	var domains []shared_types.Domain
	err := s.getDB().NewSelect().Model(&domains).
		Where("type = ?", "custom").
		Where("organization_id = ?", orgID).
		Where("deleted_at IS NULL").
		Scan(s.Ctx)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: GetCustomDomainsByOrg: %v", err), ctxStr)
		return nil, err
	}
	s.storageLog(logger.Debug, "storage: GetCustomDomainsByOrg ok", fmt.Sprintf("%s count=%d", ctxStr, len(domains)))
	return domains, nil
}

func (s *DomainStorage) GetCustomDomainByID(id uuid.UUID, orgID uuid.UUID) (*shared_types.Domain, error) {
	ctxStr := fmt.Sprintf("id=%s org_id=%s", id, orgID)
	s.storageLog(logger.Debug, "storage: GetCustomDomainByID", ctxStr)

	var domain shared_types.Domain
	err := s.getDB().NewSelect().Model(&domain).
		Where("id = ?", id).
		Where("organization_id = ?", orgID).
		Where("type = ?", "custom").
		Where("deleted_at IS NULL").
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.storageLog(logger.Debug, "storage: GetCustomDomainByID not found", ctxStr)
			return nil, types.ErrCustomDomainNotFound
		}
		s.storageLog(logger.Error, fmt.Sprintf("storage: GetCustomDomainByID: %v", err), ctxStr)
		return nil, err
	}
	return &domain, nil
}

func (s *DomainStorage) GetCustomDomainByName(name string) (*shared_types.Domain, error) {
	ctxStr := fmt.Sprintf("name=%s", name)
	s.storageLog(logger.Debug, "storage: GetCustomDomainByName", ctxStr)

	var domain shared_types.Domain
	err := s.getDB().NewSelect().Model(&domain).
		Where("name = ?", name).
		Where("type = ?", "custom").
		Where("deleted_at IS NULL").
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.storageLog(logger.Debug, "storage: GetCustomDomainByName not found", ctxStr)
			return nil, nil
		}
		s.storageLog(logger.Error, fmt.Sprintf("storage: GetCustomDomainByName: %v", err), ctxStr)
		return nil, err
	}
	return &domain, nil
}

func (s *DomainStorage) UpdateCustomDomainStatus(id uuid.UUID, status string) error {
	ctxStr := fmt.Sprintf("id=%s status=%s", id, status)
	s.storageLog(logger.Debug, "storage: UpdateCustomDomainStatus", ctxStr)

	_, err := s.getDB().NewUpdate().
		Model((*shared_types.Domain)(nil)).
		Set("status = ?", status).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(s.Ctx)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: UpdateCustomDomainStatus: %v", err), ctxStr)
		return err
	}
	s.storageLog(logger.Info, "storage: UpdateCustomDomainStatus ok", ctxStr)
	return nil
}

func (s *DomainStorage) UpdateCustomDomainVerification(id uuid.UUID, status string, dnsProvider *string) error {
	dnsStr := ""
	if dnsProvider != nil {
		dnsStr = *dnsProvider
	}
	ctxStr := fmt.Sprintf("id=%s status=%s dns_provider=%s", id, status, dnsStr)
	s.storageLog(logger.Debug, "storage: UpdateCustomDomainVerification", ctxStr)

	_, err := s.getDB().NewUpdate().
		Model((*shared_types.Domain)(nil)).
		Set("status = ?", status).
		Set("dns_provider = ?", dnsProvider).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", id).
		Where("deleted_at IS NULL").
		Exec(s.Ctx)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: UpdateCustomDomainVerification: %v", err), ctxStr)
		return err
	}
	s.storageLog(logger.Info, "storage: UpdateCustomDomainVerification ok", ctxStr)
	return nil
}

func (s *DomainStorage) DeleteCustomDomain(id uuid.UUID) error {
	ctxStr := fmt.Sprintf("id=%s", id)
	s.storageLog(logger.Debug, "storage: DeleteCustomDomain", ctxStr)

	result, err := s.getDB().NewDelete().
		Model((*shared_types.Domain)(nil)).
		Where("id = ?", id).
		Exec(s.Ctx)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: DeleteCustomDomain: %v", err), ctxStr)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: DeleteCustomDomain RowsAffected: %v", err), ctxStr)
		return err
	}

	if rowsAffected == 0 {
		s.storageLog(logger.Debug, "storage: DeleteCustomDomain no row deleted", ctxStr)
		return types.ErrCustomDomainNotFound
	}

	s.storageLog(logger.Info, "storage: DeleteCustomDomain ok", ctxStr)
	return nil
}
