package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/nixopus/nixopus/api/internal/features/trail/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

// TrailRepository defines the interface for trail storage operations.
type TrailRepository interface {
	GetActiveProvisionByUserAndOrg(userID, orgID string) (*types.UserProvisionDetails, error)
	CountActiveProvisions() (int, error)
	CreateActiveUserProvision(details *types.UserProvisionDetails) error
	GetUserProvisionDetailsByID(sessionID string) (*types.UserProvisionDetails, error)
	GetUserProvisionStatus(userID string) (types.UserProvisionStatus, error)
	UpdateUserProvisionStatus(userID string, status types.UserProvisionStatus) error
	UpdateUserProvisionDetailsWithError(sessionID string, errorMsg string) error
	UpdateUserProvisionDetailsStep(sessionID string, step types.ProvisionStep) error
	GetUserByID(userID string) (*shared_types.User, error)
	IsSubdomainTaken(subdomain string) (bool, error)
	GetCompletedProvisionByUserID(userID string) (*types.UserProvisionDetails, error)
	SelectBestServer(vcpus, memMB, diskGB int) (string, error)
}

// TrailStorage implements TrailRepository using Bun ORM.
type TrailStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

// NewTrailStorage creates a new TrailStorage instance.
func NewTrailStorage(db *bun.DB, ctx context.Context) *TrailStorage {
	return &TrailStorage{
		DB:  db,
		Ctx: ctx,
	}
}

// GetActiveProvisionByUserAndOrg retrieves an active (non-failed) provision for a user and organization.
//
// Parameters:
//   - userID: the UUID of the user
//   - orgID: the UUID of the organization
//
// Returns:
//   - *types.UserProvisionDetails: the active provision if found, nil otherwise
//   - error: database error if query fails
func (s *TrailStorage) GetActiveProvisionByUserAndOrg(userID, orgID string) (*types.UserProvisionDetails, error) {
	var provision types.UserProvisionDetails

	err := s.DB.NewSelect().
		Model(&provision).
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Where("step IS NOT NULL AND step != ?", types.ProvisionStepCompleted).
		Order("created_at DESC").
		Limit(1).
		Scan(s.Ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get active provision: %w", err)
	}

	return &provision, nil
}

// CountActiveProvisions counts the total number of active (non-completed) provisions.
//
// Returns:
//   - int: the count of active provisions
//   - error: database error if query fails
func (s *TrailStorage) CountActiveProvisions() (int, error) {
	count, err := s.DB.NewSelect().
		Model((*types.UserProvisionDetails)(nil)).
		Where("step IS NOT NULL AND step != ?", types.ProvisionStepCompleted).
		Count(s.Ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count active provisions: %w", err)
	}

	return count, nil
}

// CreateActiveUserProvision creates a new provision record atomically.
// Uses database constraints to prevent race conditions.
//
// Parameters:
//   - details: the provision details to create
//
// Returns:
//   - error: database error if insert fails
func (s *TrailStorage) CreateActiveUserProvision(details *types.UserProvisionDetails) error {
	tx, err := s.DB.BeginTx(s.Ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.NewInsert().Model(details).Exec(s.Ctx)
	if err != nil {
		return fmt.Errorf("failed to create provision: %w", err)
	}

	return tx.Commit()
}

// GetUserProvisionDetailsByID retrieves provision details by session ID.
//
// Parameters:
//   - sessionID: the UUID of the provision session
//
// Returns:
//   - *types.UserProvisionDetails: the provision details if found, nil otherwise
//   - error: database error if query fails
func (s *TrailStorage) GetUserProvisionDetailsByID(sessionID string) (*types.UserProvisionDetails, error) {
	var provision types.UserProvisionDetails

	err := s.DB.NewSelect().
		Model(&provision).
		Where("id = ?", sessionID).
		Scan(s.Ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get provision details: %w", err)
	}

	return &provision, nil
}

// GetUserProvisionStatus retrieves the provision status for a user.
//
// Parameters:
//   - userID: the UUID of the user
//
// Returns:
//   - types.UserProvisionStatus: the user's provision status
//   - error: database error if query fails
func (s *TrailStorage) GetUserProvisionStatus(userID string) (types.UserProvisionStatus, error) {
	var user shared_types.User

	err := s.DB.NewSelect().
		Model(&user).
		Column("provision_status").
		Where("id = ?", userID).
		Scan(s.Ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return types.UserProvisionStatusNotStarted, nil
		}
		return types.UserProvisionStatusNotStarted, fmt.Errorf("failed to get user provision status: %w", err)
	}

	if user.ProvisionStatus == nil {
		return types.UserProvisionStatusNotStarted, nil
	}

	return types.UserProvisionStatus(*user.ProvisionStatus), nil
}

// UpdateUserProvisionStatus updates the provision status for a user.
//
// Parameters:
//   - userID: the UUID of the user
//   - status: the new provision status
//
// Returns:
//   - error: database error if update fails
func (s *TrailStorage) UpdateUserProvisionStatus(userID string, status types.UserProvisionStatus) error {
	statusStr := string(status)
	_, err := s.DB.NewUpdate().
		Model((*shared_types.User)(nil)).
		Set("provision_status = ?", statusStr).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", userID).
		Exec(s.Ctx)

	if err != nil {
		return fmt.Errorf("failed to update user provision status: %w", err)
	}

	return nil
}

// UpdateUserProvisionDetailsWithError updates a provision record with an error message.
//
// Parameters:
//   - sessionID: the UUID of the provision session
//   - errorMsg: the error message to store
//
// Returns:
//   - error: database error if update fails
func (s *TrailStorage) UpdateUserProvisionDetailsWithError(sessionID string, errorMsg string) error {
	_, err := s.DB.NewUpdate().
		Model((*types.UserProvisionDetails)(nil)).
		Set("error = ?", errorMsg).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", sessionID).
		Exec(s.Ctx)

	if err != nil {
		return fmt.Errorf("failed to update provision details with error: %w", err)
	}

	return nil
}

// UpdateUserProvisionDetailsStep updates the step of a provision record.
//
// Parameters:
//   - sessionID: the UUID of the provision session
//   - step: the new provision step
//
// Returns:
//   - error: database error if update fails
func (s *TrailStorage) UpdateUserProvisionDetailsStep(sessionID string, step types.ProvisionStep) error {
	_, err := s.DB.NewUpdate().
		Model((*types.UserProvisionDetails)(nil)).
		Set("step = ?", step).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", sessionID).
		Exec(s.Ctx)

	if err != nil {
		return fmt.Errorf("failed to update provision step: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by their ID.
//
// Parameters:
//   - userID: the UUID of the user
//
// Returns:
//   - *shared_types.User: the user if found
//   - error: database error if query fails
func (s *TrailStorage) GetUserByID(userID string) (*shared_types.User, error) {
	var user shared_types.User

	err := s.DB.NewSelect().
		Model(&user).
		Where("id = ?", userID).
		Scan(s.Ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *TrailStorage) GetCompletedProvisionByUserID(userID string) (*types.UserProvisionDetails, error) {
	var provision types.UserProvisionDetails

	err := s.DB.NewSelect().
		Model(&provision).
		Where("upd.user_id = ?", userID).
		Where("upd.step = ?", types.ProvisionStepCompleted).
		Order("upd.created_at DESC").
		Limit(1).
		Scan(s.Ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get completed provision: %w", err)
	}

	return &provision, nil
}

// IsSubdomainTaken checks if a subdomain is already in use.
//
// Parameters:
//   - subdomain: the subdomain to check
//
// Returns:
//   - bool: true if subdomain is taken, false otherwise
//   - error: database error if query fails
func (s *TrailStorage) IsSubdomainTaken(subdomain string) (bool, error) {
	exists, err := s.DB.NewSelect().
		Model((*types.UserProvisionDetails)(nil)).
		Where("subdomain = ?", subdomain).
		Exists(s.Ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check subdomain: %w", err)
	}

	return exists, nil
}
