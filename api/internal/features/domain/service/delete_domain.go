package service

import (
	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
)

// DeleteDomain deletes an existing domain in the application.
//
// It takes a domain ID and deletes the associated domain in the storage layer.
// If the domain does not exist, it returns ErrDomainNotFound.
// If any error occurs while deleting the domain, it returns the error.
func (s *DomainsService) DeleteDomain(domainID string) error {
	s.logger.Log(logger.Debug, "delete domain request received: domain_id=%s\n", domainID)

	_, err := uuid.Parse(domainID)
	if err != nil {
		return types.ErrInvalidDomainID
	}

	tx, err := s.storage.BeginTx()
	if err != nil {
		s.logger.Log(logger.Error, "failed to start transaction", err.Error())
		return types.ErrFailedToDeleteDomain
	}
	defer tx.Rollback()

	txStorage := s.storage.WithTx(tx)

	existing_domain, err := txStorage.GetDomain(domainID)
	if err != nil {
		s.logger.Log(logger.Error, "error while retrieving domain", err.Error())
		return err
	}

	if existing_domain == nil {
		s.logger.Log(logger.Error, "domain not found", domainID)
		return types.ErrDomainNotFound
	}

	if err := txStorage.DeleteDomain(existing_domain); err != nil {
		s.logger.Log(logger.Error, "error while deleting domain", err.Error())
		return err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Log(logger.Error, "failed to commit transaction", err.Error())
		return types.ErrFailedToDeleteDomain
	}

	return nil
}
