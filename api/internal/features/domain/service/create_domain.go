package service

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/types"
	"github.com/raghavyuva/nixopus-api/internal/features/domain/validation"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// CreateDomain creates a new domain in the application.
//
// It takes a CreateDomainRequest, which contains the domain name, and a user ID.
// The user ID is used to associate the domain with a user.
//
// It returns a CreateDomainResponse, which is always empty, and an error.
// The error is either ErrDomainAlreadyExists, or any error that occurred
// while creating the domain in the storage layer.
func (s *DomainsService) CreateDomain(req types.CreateDomainRequest, userID string) (types.CreateDomainResponse, error) {
	s.logger.Log(logger.Info, "create domain request received", fmt.Sprintf("domain_name=%s, user_id=%s", req.Name, userID))

	_, err := uuid.Parse(userID)
	if err != nil {
		s.logger.Log(logger.Error, "invalid user id", fmt.Sprintf("user_id=%s", userID))
		return types.CreateDomainResponse{}, types.ErrInvalidUserID
	}

	if userID == "" {
		s.logger.Log(logger.Error, "invalid user id", fmt.Sprintf("user_id=%s", userID))
		return types.CreateDomainResponse{}, types.ErrInvalidUserID
	}

	validator := validation.NewValidator(s.storage)
	if err := validator.ValidateCreateDomainRequest(req); err != nil {
		return types.CreateDomainResponse{}, err
	}

	if err := validator.ValidateDomainBelongsToServer(req.Name); err != nil {
		s.logger.Log(logger.Error, "domain does not belong to server", fmt.Sprintf("domain_name=%s", req.Name))
		return types.CreateDomainResponse{}, err
	}

	// Verify organization exists in local database (for foreign key constraints)
	var org shared_types.Organization
	err = s.store.DB.NewSelect().
		Model(&org).
		Where("id = ?", req.OrganizationID.String()).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Log(logger.Error, "organization not found", req.OrganizationID.String())
			return types.CreateDomainResponse{}, fmt.Errorf("organization not found")
		}
		s.logger.Log(logger.Error, "error while retrieving organization", err.Error())
		return types.CreateDomainResponse{}, fmt.Errorf("organization not found")
	}
	if org.ID == uuid.Nil {
		s.logger.Log(logger.Error, "organization not found", req.OrganizationID.String())
		return types.CreateDomainResponse{}, fmt.Errorf("organization not found")
	}

	tx, err := s.storage.BeginTx()
	if err != nil {
		s.logger.Log(logger.Error, "failed to start transaction", err.Error())
		return types.CreateDomainResponse{}, types.ErrFailedToCreateDomain
	}
	defer tx.Rollback()

	txStorage := s.storage.WithTx(tx)

	existing_domain, err := txStorage.GetDomainByName(req.Name, req.OrganizationID)
	if err != nil {
		s.logger.Log(logger.Error, "error while retrieving domain", err.Error())
		return types.CreateDomainResponse{}, err
	}

	if existing_domain != nil {
		s.logger.Log(logger.Error, "domain already exists", fmt.Sprintf("domain_name=%s", req.Name))
		return types.CreateDomainResponse{}, types.ErrDomainAlreadyExists
	}

	domain := &shared_types.Domain{
		ID:             uuid.New(),
		UserID:         uuid.MustParse(userID),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		DeletedAt:      nil,
		Name:           req.Name,
		OrganizationID: req.OrganizationID,
	}

	if err := txStorage.CreateDomain(domain); err != nil {
		s.logger.Log(logger.Error, "error while creating domain", err.Error())
		return types.CreateDomainResponse{}, err
	}

	if err := tx.Commit(); err != nil {
		s.logger.Log(logger.Error, "failed to commit transaction", err.Error())
		return types.CreateDomainResponse{}, types.ErrFailedToCreateDomain
	}

	return types.CreateDomainResponse{ID: domain.ID.String()}, nil
}
