package storage

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/trail/types"
)

type ExpiredTrialUser struct {
	ProvisionID      uuid.UUID  `bun:"provision_id"`
	UserID           uuid.UUID  `bun:"user_id"`
	OrganizationID   uuid.UUID  `bun:"organization_id"`
	ServerID         *uuid.UUID `bun:"server_id"`
	LXDContainerName *string    `bun:"lxd_container_name"`
	Subdomain        *string    `bun:"subdomain"`
	Email            string     `bun:"email"`
	Name             string     `bun:"name"`
}

func (s *TrailStorage) GetExpiredTrialUsers(ctx context.Context, trialPeriodDays int) ([]ExpiredTrialUser, error) {
	var users []ExpiredTrialUser

	err := s.DB.NewRaw(`
		SELECT
			upd.id AS provision_id,
			upd.user_id,
			upd.organization_id,
			upd.server_id,
			upd.lxd_container_name,
			upd.subdomain,
			u.email,
			COALESCE(u.name, '') AS name
		FROM user_provision_details upd
		JOIN "user" u ON u.id = upd.user_id
		WHERE u.provision_status = ?
			AND upd.step = ?
			AND upd.created_at + make_interval(days => ?) < now()
			AND NOT EXISTS (
				SELECT 1 FROM org_machine_billing omb
				WHERE omb.organization_id = upd.organization_id
			)
			AND NOT EXISTS (
				SELECT 1 FROM applications app
				WHERE app.organization_id = upd.organization_id
			)
	`, string(types.UserProvisionStatusCompleted), string(types.ProvisionStepCompleted), trialPeriodDays).Scan(ctx, &users)

	if err != nil {
		return nil, fmt.Errorf("failed to query expired trial users: %w", err)
	}

	return users, nil
}

func (s *TrailStorage) HasMachineBilling(ctx context.Context, orgID uuid.UUID) (bool, error) {
	exists, err := s.DB.NewSelect().
		TableExpr("org_machine_billing").
		Where("organization_id = ?", orgID).
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check machine billing: %w", err)
	}
	return exists, nil
}

func (s *TrailStorage) HasApplications(ctx context.Context, orgID uuid.UUID) (bool, error) {
	exists, err := s.DB.NewSelect().
		TableExpr("applications").
		Where("organization_id = ?", orgID).
		Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check applications: %w", err)
	}
	return exists, nil
}

func (s *TrailStorage) DeleteProvisionAndResetStatus(ctx context.Context, provisionID, userID uuid.UUID) error {
	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.NewDelete().
		Model((*types.UserProvisionDetails)(nil)).
		Where("id = ?", provisionID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete provision details: %w", err)
	}

	statusStr := string(types.UserProvisionStatusPending)
	_, err = tx.NewRaw(
		`UPDATE "user" SET provision_status = ?, updated_at = now() WHERE id = ?`,
		statusStr, userID,
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to reset provision status: %w", err)
	}

	return tx.Commit()
}
