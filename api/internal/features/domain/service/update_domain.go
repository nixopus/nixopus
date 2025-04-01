package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// UpdateDomain updates an existing domain in the application.
//
// It takes a domain ID, a userID and a new name and updates the domain in the storage layer.
// If the domain does not exist, it returns ErrDomainNotFound.
// If any error occurs while updating the domain, it returns the error.
// UpdateDomain updates a domain's name.
//
// It takes a new name, user ID, and domain ID, then updates the domain in the storage layer.
// Returns the updated domain and any error that occurred.
func (s *DomainsService) UpdateDomain(newName, userID, domainID string) (*shared_types.Domain, error) {
	s.logger.Log(logger.Debug, fmt.Sprintf("update domain request received: domain_id=%s, user_id=%s, new_name=%s", domainID, userID, newName), "")

	domainUUID, err := uuid.Parse(domainID)
	if err != nil {
		return nil, types.ErrInvalidDomainID
	}

	if domainUUID == uuid.Nil {
		return nil, types.ErrInvalidDomainID
	}

	tx, err := s.storage.BeginTx()
	if err != nil {
		s.logger.Log(logger.Error, "failed to start transaction", err.Error())
		return nil, types.ErrFailedToUpdateDomain
	}
	defer tx.Rollback()

	txStorage := s.storage.WithTx(tx)

	existing, err := txStorage.GetDomain(domainID)
	if err != nil {
		s.logger.Log(logger.Error, "error while retrieving domain", err.Error())
		return nil, err
	}

	if existing == nil {
		s.logger.Log(logger.Error, "domain not found", domainID)
		return nil, types.ErrDomainNotFound
	}

	validator := validation.NewValidator(s.storage)
	if err := validator.ValidateDomainBelongsToServer(newName); err != nil {
		s.logger.Log(logger.Error, "domain does not belong to server", fmt.Sprintf("domain_name=%s", newName))
		return nil, err
	}

	err = txStorage.UpdateDomain(domainID, newName)
	if err != nil {
		s.logger.Log(logger.Error, "error while updating domain", err.Error())
		return nil, err
	}

	updated, err := txStorage.GetDomain(domainID)
	if err != nil {
		s.logger.Log(logger.Error, "error while retrieving updated domain", err.Error())
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Log(logger.Error, "failed to commit transaction", err.Error())
		return nil, types.ErrFailedToUpdateDomain
	}

	return updated, nil
}
