package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"

	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
)

// TrailRepository defines the interface for trial storage operations.
type TrailRepository interface {
	GetActiveProvisionByUserAndOrg(userID, orgID string) (*shared_types.UserProvisionDetails, error)
	CountActiveProvisions() (int, error)
	CreateActiveUserProvision(details *shared_types.UserProvisionDetails) error
	GetUserProvisionDetailsByID(sessionID string) (*shared_types.UserProvisionDetails, error)
	GetUserProvisionStatus(userID string) (machine_types.UserProvisionStatus, error)
	UpdateUserProvisionStatus(userID string, status machine_types.UserProvisionStatus) error
	UpdateUserProvisionDetailsWithError(sessionID string, errorMsg string) error
	UpdateUserProvisionDetailsStep(sessionID string, step shared_types.ProvisionStep) error
	GetUserByID(userID string) (*shared_types.User, error)
	IsSubdomainTaken(subdomain string) (bool, error)
	GetCompletedProvisionByUserID(userID string) (*shared_types.UserProvisionDetails, error)
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

func (s *TrailStorage) GetActiveProvisionByUserAndOrg(userID, orgID string) (*shared_types.UserProvisionDetails, error) {
	var provision shared_types.UserProvisionDetails

	err := s.DB.NewSelect().
		Model(&provision).
		Where("user_id = ? AND organization_id = ?", userID, orgID).
		Where("step IS NOT NULL AND step != ?", shared_types.ProvisionStepCompleted).
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

func (s *TrailStorage) CountActiveProvisions() (int, error) {
	count, err := s.DB.NewSelect().
		Model((*shared_types.UserProvisionDetails)(nil)).
		Where("step IS NOT NULL AND step != ?", shared_types.ProvisionStepCompleted).
		Count(s.Ctx)

	if err != nil {
		return 0, fmt.Errorf("failed to count active provisions: %w", err)
	}

	return count, nil
}

func (s *TrailStorage) CreateActiveUserProvision(details *shared_types.UserProvisionDetails) error {
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

func (s *TrailStorage) GetUserProvisionDetailsByID(sessionID string) (*shared_types.UserProvisionDetails, error) {
	var provision shared_types.UserProvisionDetails

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

func (s *TrailStorage) GetUserProvisionStatus(userID string) (machine_types.UserProvisionStatus, error) {
	var user shared_types.User

	err := s.DB.NewSelect().
		Model(&user).
		Column("provision_status").
		Where("id = ?", userID).
		Scan(s.Ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return machine_types.UserProvisionStatusPending, nil
		}
		return machine_types.UserProvisionStatusPending, fmt.Errorf("failed to get user provision status: %w", err)
	}

	if user.ProvisionStatus == nil {
		return machine_types.UserProvisionStatusPending, nil
	}

	return machine_types.UserProvisionStatus(*user.ProvisionStatus), nil
}

func (s *TrailStorage) UpdateUserProvisionStatus(userID string, status machine_types.UserProvisionStatus) error {
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

func (s *TrailStorage) UpdateUserProvisionDetailsWithError(sessionID string, errorMsg string) error {
	_, err := s.DB.NewUpdate().
		Model((*shared_types.UserProvisionDetails)(nil)).
		Set("error = ?", errorMsg).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", sessionID).
		Exec(s.Ctx)

	if err != nil {
		return fmt.Errorf("failed to update provision details with error: %w", err)
	}

	return nil
}

func (s *TrailStorage) UpdateUserProvisionDetailsStep(sessionID string, step shared_types.ProvisionStep) error {
	_, err := s.DB.NewUpdate().
		Model((*shared_types.UserProvisionDetails)(nil)).
		Set("step = ?", step).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", sessionID).
		Exec(s.Ctx)

	if err != nil {
		return fmt.Errorf("failed to update provision step: %w", err)
	}

	return nil
}

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

func (s *TrailStorage) GetCompletedProvisionByUserID(userID string) (*shared_types.UserProvisionDetails, error) {
	var provision shared_types.UserProvisionDetails

	err := s.DB.NewSelect().
		Model(&provision).
		Where("upd.user_id = ?", userID).
		Where("upd.step = ?", shared_types.ProvisionStepCompleted).
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

func (s *TrailStorage) IsSubdomainTaken(subdomain string) (bool, error) {
	exists, err := s.DB.NewSelect().
		Model((*shared_types.UserProvisionDetails)(nil)).
		Where("subdomain = ?", subdomain).
		Exists(s.Ctx)

	if err != nil {
		return false, fmt.Errorf("failed to check subdomain: %w", err)
	}

	return exists, nil
}

// ---- scheduler helpers ----

type serverCandidate struct {
	bun.BaseModel `bun:"table:infra_servers,alias:isr"`

	ID         uuid.UUID `bun:"id"`
	MaxVCPUs   int       `bun:"max_vcpus"`
	MaxMemMB   int       `bun:"max_memory_mb"`
	MaxDiskGB  int       `bun:"max_disk_gb"`
	UsedVCPUs  int       `bun:"used_vcpus"`
	UsedMemMB  int       `bun:"used_mem"`
	UsedDiskGB int       `bun:"used_disk"`
}

var trialActiveProvisionStatuses = []string{"provisioning", "completed"}

// SelectBestServer picks the active infra server with the most remaining capacity.
// Returns an empty string (not an error) when no infra_servers are registered.
func (s *TrailStorage) SelectBestServer(vcpus, memMB, diskGB int) (string, error) {
	var candidates []serverCandidate

	err := s.DB.NewSelect().
		TableExpr("infra_servers AS isr").
		ColumnExpr("isr.id").
		ColumnExpr("isr.max_vcpus").
		ColumnExpr("isr.max_memory_mb").
		ColumnExpr("isr.max_disk_gb").
		ColumnExpr("COALESCE(SUM(upd.vcpu_count), 0) + COALESCE(pv.pool_vcpus, 0) AS used_vcpus").
		ColumnExpr("COALESCE(SUM(upd.memory_mb), 0) + COALESCE(pv.pool_mem, 0) AS used_mem").
		ColumnExpr("COALESCE(SUM(upd.disk_size_gb), 0) + COALESCE(pv.pool_disk, 0) AS used_disk").
		Join(`LEFT JOIN user_provision_details AS upd ON upd.server_id = isr.id`+
			` AND EXISTS (SELECT 1 FROM "user" u WHERE u.id = upd.user_id AND u.provision_status IN (?))`,
			bun.In(trialActiveProvisionStatuses)).
		Join(`LEFT JOIN (`+
			`SELECT vp.server_id, `+
			`SUM(vp.vcpu_count) AS pool_vcpus, `+
			`SUM(vp.memory_mb) AS pool_mem, `+
			`SUM(vp.disk_size_gb) AS pool_disk `+
			`FROM vm_pool AS vp `+
			`WHERE vp.status IN ('warm', 'claiming') `+
			`GROUP BY vp.server_id`+
			`) AS pv ON pv.server_id = isr.id`).
		Where("isr.status = ?", "active").
		GroupExpr("isr.id, pv.pool_vcpus, pv.pool_mem, pv.pool_disk").
		Scan(s.Ctx, &candidates)

	if err != nil {
		return "", fmt.Errorf("scheduler query failed: %w", err)
	}

	if len(candidates) == 0 {
		return "", nil
	}

	return pickBestTrialServer(candidates, vcpus, memMB, diskGB)
}

func pickBestTrialServer(candidates []serverCandidate, vcpus, memMB, diskGB int) (string, error) {
	bestIdx := -1
	bestMinHeadroom := -1.0

	for i, c := range candidates {
		freeVCPU := c.MaxVCPUs - c.UsedVCPUs
		freeMem := c.MaxMemMB - c.UsedMemMB
		freeDisk := c.MaxDiskGB - c.UsedDiskGB

		if freeVCPU < vcpus || freeMem < memMB || freeDisk < diskGB {
			continue
		}

		hCPU := float64(freeVCPU-vcpus) / float64(max(c.MaxVCPUs, 1))
		hMem := float64(freeMem-memMB) / float64(max(c.MaxMemMB, 1))
		hDisk := float64(freeDisk-diskGB) / float64(max(c.MaxDiskGB, 1))

		minH := hCPU
		if hMem < minH {
			minH = hMem
		}
		if hDisk < minH {
			minH = hDisk
		}

		if minH > bestMinHeadroom {
			bestMinHeadroom = minH
			bestIdx = i
		}
	}

	if bestIdx < 0 {
		return "", fmt.Errorf("no server has enough capacity for the requested resources (vcpus=%d mem=%dMB disk=%dGB)", vcpus, memMB, diskGB)
	}

	return candidates[bestIdx].ID.String(), nil
}

// ---- trial_expiry helpers ----

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
	`, string(machine_types.UserProvisionStatusCompleted), string(shared_types.ProvisionStepCompleted), trialPeriodDays).Scan(ctx, &users)

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
		Model((*shared_types.UserProvisionDetails)(nil)).
		Where("id = ?", provisionID).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete provision details: %w", err)
	}

	statusStr := string(machine_types.UserProvisionStatusPending)
	_, err = tx.NewRaw(
		`UPDATE "user" SET provision_status = ?, updated_at = now() WHERE id = ?`,
		statusStr, userID,
	).Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to reset provision status: %w", err)
	}

	return tx.Commit()
}
