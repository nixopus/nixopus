package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

type DomainStorage struct {
	DB     *bun.DB
	Ctx    context.Context
	Logger *logger.Logger // optional; when nil, storage emits no logs
}

func (s *DomainStorage) storageLog(sev logger.Severity, msg, data string) {
	if s.Logger == nil {
		return
	}
	s.Logger.Log(sev, msg, data)
}

type DomainStorageInterface interface {
	GetDomains(OrganizationID string, UserID uuid.UUID) ([]shared_types.Domain, error)
	CreateCustomDomain(domain *shared_types.Domain) error
	GetCustomDomainsByOrg(orgID uuid.UUID) ([]shared_types.Domain, error)
	GetCustomDomainByID(id uuid.UUID, orgID uuid.UUID) (*shared_types.Domain, error)
	GetCustomDomainByName(name string) (*shared_types.Domain, error)
	UpdateCustomDomainStatus(id uuid.UUID, status string) error
	UpdateCustomDomainVerification(id uuid.UUID, status string, dnsProvider *string) error
	DeleteCustomDomain(id uuid.UUID) error

	GetDefaultSSHKeyByOrg(orgID uuid.UUID) (*shared_types.SSHKey, error)
	GetProvisionDetailsBySSHKeyAndOrg(sshKeyID, orgID uuid.UUID) (*shared_types.UserProvisionDetails, error)
	GetProvisionDetailsBySubdomain(subdomain string) (*shared_types.UserProvisionDetails, error)
}

func (s *DomainStorage) getDB() bun.IDB {
	return s.DB
}

func (s *DomainStorage) GetDomains(OrganizationID string, UserID uuid.UUID) ([]shared_types.Domain, error) {
	ctxStr := fmt.Sprintf("org_id=%s user_id=%s", OrganizationID, UserID)
	s.storageLog(logger.Debug, "storage: GetDomains", ctxStr)

	var domains []shared_types.Domain
	err := s.getDB().NewSelect().Model(&domains).
		Where("organization_id = ? AND deleted_at IS NULL", OrganizationID).
		Scan(s.Ctx)
	if err != nil {
		s.storageLog(logger.Error, fmt.Sprintf("storage: GetDomains: %v", err), ctxStr)
		return nil, err
	}
	s.storageLog(logger.Debug, "storage: GetDomains ok", fmt.Sprintf("%s count=%d", ctxStr, len(domains)))
	return domains, nil
}
