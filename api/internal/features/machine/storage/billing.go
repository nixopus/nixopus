package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/uptrace/bun"
)

type BillingStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

func NewBillingStorage(db *bun.DB, ctx context.Context) *BillingStorage {
	return &BillingStorage{DB: db, Ctx: ctx}
}

func (s *BillingStorage) ListActivePlans() ([]types.MachinePlan, error) {
	var plans []types.MachinePlan
	err := s.DB.NewSelect().
		Model(&plans).
		Where("is_active = ?", true).
		OrderExpr("monthly_cost_cents ASC").
		Scan(s.Ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list machine plans: %w", err)
	}
	return plans, nil
}

func (s *BillingStorage) GetPlanByTier(tier string) (*types.MachinePlan, error) {
	var plan types.MachinePlan
	err := s.DB.NewSelect().
		Model(&plan).
		Where("tier = ? AND is_active = ?", tier, true).
		Limit(1).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get plan by tier: %w", err)
	}
	return &plan, nil
}

func (s *BillingStorage) GetPlanByID(planID uuid.UUID) (*types.MachinePlan, error) {
	var plan types.MachinePlan
	err := s.DB.NewSelect().
		Model(&plan).
		Where("id = ?", planID).
		Limit(1).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get plan by id: %w", err)
	}
	return &plan, nil
}

func (s *BillingStorage) GetBillingByOrgID(orgID uuid.UUID) (*types.OrgMachineBilling, error) {
	var billing types.OrgMachineBilling
	err := s.DB.NewSelect().
		Model(&billing).
		Where("organization_id = ?", orgID).
		Limit(1).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get billing by org: %w", err)
	}
	return &billing, nil
}

func (s *BillingStorage) GetWalletBalance(orgID uuid.UUID) (int, error) {
	var tx types.WalletTransaction
	err := s.DB.NewSelect().
		Model(&tx).
		Column("balance_after_cents").
		Where("organization_id = ?", orgID).
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(s.Ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get wallet balance: %w", err)
	}
	return tx.BalanceAfterCents, nil
}

func (s *BillingStorage) DebitWallet(orgID uuid.UUID, amountCents int, reason string, referenceID string) (bool, error) {
	bunTx, err := s.DB.BeginTx(s.Ctx, nil)
	if err != nil {
		return false, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer bunTx.Rollback()

	_, err = bunTx.ExecContext(s.Ctx, "SELECT pg_advisory_xact_lock(hashtext(?))", orgID.String())
	if err != nil {
		return false, fmt.Errorf("failed to acquire advisory lock: %w", err)
	}

	var existing types.WalletTransaction
	existsErr := bunTx.NewSelect().
		Model(&existing).
		Where("reference_id = ?", referenceID).
		Limit(1).
		Scan(s.Ctx)
	if existsErr == nil {
		return false, nil
	}
	if !errors.Is(existsErr, sql.ErrNoRows) {
		return false, fmt.Errorf("failed to check existing transaction: %w", existsErr)
	}

	var lastTx types.WalletTransaction
	balanceErr := bunTx.NewSelect().
		Model(&lastTx).
		Column("balance_after_cents").
		Where("organization_id = ?", orgID).
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(s.Ctx)

	currentBalance := 0
	if balanceErr == nil {
		currentBalance = lastTx.BalanceAfterCents
	} else if !errors.Is(balanceErr, sql.ErrNoRows) {
		return false, fmt.Errorf("failed to get current balance: %w", balanceErr)
	}

	if currentBalance < amountCents {
		return false, nil
	}

	newBalance := currentBalance - amountCents
	_, err = bunTx.NewInsert().Model(&types.WalletTransaction{
		OrganizationID:    orgID,
		AmountCents:       amountCents,
		EntryType:         "debit",
		BalanceAfterCents: newBalance,
		Reason:            &reason,
		ReferenceID:       &referenceID,
	}).Exec(s.Ctx)
	if err != nil {
		return false, fmt.Errorf("failed to insert wallet transaction: %w", err)
	}

	return true, bunTx.Commit()
}

func (s *BillingStorage) UpsertBilling(orgID uuid.UUID, planID uuid.UUID, periodStart, periodEnd time.Time) error {
	existing, err := s.GetBillingByOrgID(orgID)
	if err != nil {
		return err
	}

	now := time.Now()

	if existing != nil {
		_, err = s.DB.NewUpdate().
			Model((*types.OrgMachineBilling)(nil)).
			Set("machine_plan_id = ?", planID).
			Set("status = ?", types.MachineBillingStatusActive).
			Set("current_period_start = ?", periodStart).
			Set("current_period_end = ?", periodEnd).
			Set("last_charged_at = ?", now).
			Set("grace_deadline = NULL").
			Set("updated_at = ?", now).
			Where("id = ?", existing.ID).
			Exec(s.Ctx)
		return err
	}

	_, err = s.DB.NewInsert().Model(&types.OrgMachineBilling{
		OrganizationID:     orgID,
		MachinePlanID:      planID,
		Status:             types.MachineBillingStatusActive,
		CurrentPeriodStart: periodStart,
		CurrentPeriodEnd:   periodEnd,
		LastChargedAt:      &now,
	}).Exec(s.Ctx)
	return err
}

type SSHKey struct {
	bun.BaseModel  `bun:"table:ssh_keys,alias:sk" swaggerignore:"true"`
	ID             uuid.UUID `bun:"id,pk,type:uuid"`
	OrganizationID uuid.UUID `bun:"organization_id,type:uuid"`
	IsActive       bool      `bun:"is_active"`
}

func (s *BillingStorage) HasActiveSSHKey(orgID uuid.UUID) (bool, error) {
	exists, err := s.DB.NewSelect().
		Model((*SSHKey)(nil)).
		Where("organization_id = ? AND is_active = ?", orgID, true).
		Exists(s.Ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check ssh key: %w", err)
	}
	return exists, nil
}

func (s *BillingStorage) HasTrialWithoutActiveBilling(orgID uuid.UUID) (bool, error) {
	exists, err := s.DB.NewSelect().
		TableExpr("user_provision_details AS upd").
		Where("upd.organization_id = ?", orgID).
		Where("upd.type = 'trial'").
		Where("NOT EXISTS (SELECT 1 FROM org_machine_billing AS omb WHERE omb.organization_id = upd.organization_id AND omb.status = 'active')").
		Exists(s.Ctx)
	return exists, err
}

type BillingWithPlan struct {
	Billing types.OrgMachineBilling `bun:"embed:omb__"`
	Plan    types.MachinePlan       `bun:"embed:mp__"`
}

func (s *BillingStorage) GetDueBillings(ctx context.Context) ([]BillingWithPlan, error) {
	var rows []BillingWithPlan
	err := s.DB.NewSelect().
		TableExpr("org_machine_billing AS omb").
		ColumnExpr("omb.id AS omb__id, omb.organization_id AS omb__organization_id, omb.ssh_key_id AS omb__ssh_key_id, omb.machine_plan_id AS omb__machine_plan_id, omb.status AS omb__status, omb.current_period_start AS omb__current_period_start, omb.current_period_end AS omb__current_period_end, omb.grace_deadline AS omb__grace_deadline, omb.last_charged_at AS omb__last_charged_at").
		ColumnExpr("mp.id AS mp__id, mp.tier AS mp__tier, mp.name AS mp__name, mp.ram_mb AS mp__ram_mb, mp.vcpu AS mp__vcpu, mp.storage_mb AS mp__storage_mb, mp.monthly_cost_cents AS mp__monthly_cost_cents").
		Join("INNER JOIN machine_plans AS mp ON omb.machine_plan_id = mp.id").
		Where("omb.status = ?", types.MachineBillingStatusActive).
		Where("omb.current_period_end <= NOW()").
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("failed to get due billings: %w", err)
	}
	return rows, nil
}

func (s *BillingStorage) GetGraceBillings(ctx context.Context) ([]BillingWithPlan, error) {
	var rows []BillingWithPlan
	err := s.DB.NewSelect().
		TableExpr("org_machine_billing AS omb").
		ColumnExpr("omb.id AS omb__id, omb.organization_id AS omb__organization_id, omb.ssh_key_id AS omb__ssh_key_id, omb.machine_plan_id AS omb__machine_plan_id, omb.status AS omb__status, omb.current_period_start AS omb__current_period_start, omb.current_period_end AS omb__current_period_end, omb.grace_deadline AS omb__grace_deadline, omb.last_charged_at AS omb__last_charged_at").
		ColumnExpr("mp.id AS mp__id, mp.tier AS mp__tier, mp.name AS mp__name, mp.ram_mb AS mp__ram_mb, mp.vcpu AS mp__vcpu, mp.storage_mb AS mp__storage_mb, mp.monthly_cost_cents AS mp__monthly_cost_cents").
		Join("INNER JOIN machine_plans AS mp ON omb.machine_plan_id = mp.id").
		Where("omb.status = ?", types.MachineBillingStatusGracePeriod).
		Scan(ctx, &rows)
	if err != nil {
		return nil, fmt.Errorf("failed to get grace billings: %w", err)
	}
	return rows, nil
}

func (s *BillingStorage) UpdateBillingPeriod(ctx context.Context, billingID uuid.UUID, periodStart, periodEnd time.Time) error {
	_, err := s.DB.NewUpdate().
		Model((*types.OrgMachineBilling)(nil)).
		Set("current_period_start = ?", periodStart).
		Set("current_period_end = ?", periodEnd).
		Set("last_charged_at = ?", time.Now()).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", billingID).
		Exec(ctx)
	return err
}

func (s *BillingStorage) SetGracePeriod(ctx context.Context, billingID uuid.UUID, deadline time.Time) error {
	_, err := s.DB.NewUpdate().
		Model((*types.OrgMachineBilling)(nil)).
		Set("status = ?", types.MachineBillingStatusGracePeriod).
		Set("grace_deadline = ?", deadline).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", billingID).
		Exec(ctx)
	return err
}

func (s *BillingStorage) RecoverFromGrace(ctx context.Context, billingID uuid.UUID, periodStart, periodEnd time.Time) error {
	now := time.Now()
	_, err := s.DB.NewUpdate().
		Model((*types.OrgMachineBilling)(nil)).
		Set("status = ?", types.MachineBillingStatusActive).
		Set("current_period_start = ?", periodStart).
		Set("current_period_end = ?", periodEnd).
		Set("last_charged_at = ?", now).
		Set("grace_deadline = NULL").
		Set("updated_at = ?", now).
		Where("id = ?", billingID).
		Exec(ctx)
	return err
}

func (s *BillingStorage) SuspendBilling(ctx context.Context, billingID uuid.UUID) error {
	_, err := s.DB.NewUpdate().
		Model((*types.OrgMachineBilling)(nil)).
		Set("status = ?", types.MachineBillingStatusSuspended).
		Set("updated_at = ?", time.Now()).
		Where("id = ?", billingID).
		Exec(ctx)
	return err
}

func (s *BillingStorage) DeactivateSSHKey(ctx context.Context, sshKeyID uuid.UUID) error {
	_, err := s.DB.NewUpdate().
		Model((*SSHKey)(nil)).
		Set("is_active = ?", false).
		Where("id = ?", sshKeyID).
		Exec(ctx)
	return err
}

func (s *BillingStorage) ReactivateSSHKey(ctx context.Context, sshKeyID uuid.UUID) error {
	_, err := s.DB.NewUpdate().
		Model((*SSHKey)(nil)).
		Set("is_active = ?", true).
		Where("id = ?", sshKeyID).
		Exec(ctx)
	return err
}

type UserProvisionDetail struct {
	bun.BaseModel    `bun:"table:user_provision_details,alias:upd" swaggerignore:"true"`
	UserID           uuid.UUID  `bun:"user_id,type:uuid"`
	OrganizationID   uuid.UUID  `bun:"organization_id,type:uuid"`
	LXDContainerName *string    `bun:"lxd_container_name"`
	ServerID         *uuid.UUID `bun:"server_id,type:uuid"`
}

type ProvisionInfo struct {
	UserID        uuid.UUID
	ContainerName string
	ServerID      string
}

func (s *BillingStorage) GetProvisionInfo(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (*ProvisionInfo, error) {
	var row UserProvisionDetail
	q := s.DB.NewSelect().Model(&row).Column("user_id", "lxd_container_name", "server_id")
	if serverID != nil {
		q = q.Where("server_id = ? AND organization_id = ?", *serverID, orgID)
	} else {
		q = q.Where("organization_id = ?", orgID).
			Where("upd.type != 'user_owned'")
	}
	err := q.OrderExpr("created_at DESC").Limit(1).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get provision info: %w", err)
	}
	info := &ProvisionInfo{UserID: row.UserID}
	if row.LXDContainerName != nil {
		info.ContainerName = *row.LXDContainerName
	}
	if row.ServerID != nil {
		info.ServerID = row.ServerID.String()
	}
	return info, nil
}
