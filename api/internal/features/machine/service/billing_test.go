package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- mock BillingRepository ----------

type mockBillingRepo struct {
	listActivePlansFn              func() ([]types.MachinePlan, error)
	getPlanByTierFn                func(tier string) (*types.MachinePlan, error)
	getPlanByIDFn                  func(planID uuid.UUID) (*types.MachinePlan, error)
	getBillingByOrgIDFn            func(orgID uuid.UUID) (*types.OrgMachineBilling, error)
	getWalletBalanceFn             func(orgID uuid.UUID) (int, error)
	debitWalletFn                  func(orgID uuid.UUID, amountCents int, reason string, referenceID string) (bool, error)
	upsertBillingFn                func(orgID uuid.UUID, planID uuid.UUID, periodStart, periodEnd time.Time) error
	hasActiveSSHKeyFn              func(orgID uuid.UUID) (bool, error)
	hasTrialWithoutActiveBillingFn func(orgID uuid.UUID) (bool, error)
	isServerUserOwnedFn            func(orgID uuid.UUID, serverID uuid.UUID) (bool, error)
	getProvisionInfoFn             func(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (*storage.ProvisionInfo, error)
	reactivateSSHKeyFn             func(ctx context.Context, sshKeyID uuid.UUID) error
}

func (m *mockBillingRepo) ListActivePlans() ([]types.MachinePlan, error) {
	if m.listActivePlansFn != nil {
		return m.listActivePlansFn()
	}
	return nil, nil
}
func (m *mockBillingRepo) GetPlanByTier(tier string) (*types.MachinePlan, error) {
	if m.getPlanByTierFn != nil {
		return m.getPlanByTierFn(tier)
	}
	return nil, nil
}
func (m *mockBillingRepo) GetPlanByID(planID uuid.UUID) (*types.MachinePlan, error) {
	if m.getPlanByIDFn != nil {
		return m.getPlanByIDFn(planID)
	}
	return nil, nil
}
func (m *mockBillingRepo) GetBillingByOrgID(orgID uuid.UUID) (*types.OrgMachineBilling, error) {
	if m.getBillingByOrgIDFn != nil {
		return m.getBillingByOrgIDFn(orgID)
	}
	return nil, nil
}
func (m *mockBillingRepo) GetWalletBalance(orgID uuid.UUID) (int, error) {
	if m.getWalletBalanceFn != nil {
		return m.getWalletBalanceFn(orgID)
	}
	return 0, nil
}
func (m *mockBillingRepo) DebitWallet(orgID uuid.UUID, amountCents int, reason string, referenceID string) (bool, error) {
	if m.debitWalletFn != nil {
		return m.debitWalletFn(orgID, amountCents, reason, referenceID)
	}
	return true, nil
}
func (m *mockBillingRepo) UpsertBilling(orgID uuid.UUID, planID uuid.UUID, periodStart, periodEnd time.Time) error {
	if m.upsertBillingFn != nil {
		return m.upsertBillingFn(orgID, planID, periodStart, periodEnd)
	}
	return nil
}
func (m *mockBillingRepo) HasActiveSSHKey(orgID uuid.UUID) (bool, error) {
	if m.hasActiveSSHKeyFn != nil {
		return m.hasActiveSSHKeyFn(orgID)
	}
	return false, nil
}
func (m *mockBillingRepo) HasTrialWithoutActiveBilling(orgID uuid.UUID) (bool, error) {
	if m.hasTrialWithoutActiveBillingFn != nil {
		return m.hasTrialWithoutActiveBillingFn(orgID)
	}
	return false, nil
}
func (m *mockBillingRepo) IsServerUserOwned(orgID uuid.UUID, serverID uuid.UUID) (bool, error) {
	if m.isServerUserOwnedFn != nil {
		return m.isServerUserOwnedFn(orgID, serverID)
	}
	return false, nil
}
func (m *mockBillingRepo) GetProvisionInfo(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (*storage.ProvisionInfo, error) {
	if m.getProvisionInfoFn != nil {
		return m.getProvisionInfoFn(ctx, orgID, serverID)
	}
	return nil, nil
}
func (m *mockBillingRepo) ReactivateSSHKey(ctx context.Context, sshKeyID uuid.UUID) error {
	if m.reactivateSSHKeyFn != nil {
		return m.reactivateSSHKeyFn(ctx, sshKeyID)
	}
	return nil
}

func TestNewBillingService(t *testing.T) {
	svc := NewBillingService(nil)
	assert.NotNil(t, svc)
}

func TestNewBillingServiceWith(t *testing.T) {
	repo := &mockBillingRepo{}
	svc := NewBillingServiceWith(repo)
	assert.NotNil(t, svc)
}

func TestPlanToResponse(t *testing.T) {
	id := uuid.New()
	plan := types.MachinePlan{
		ID:               id,
		Tier:             "starter",
		Name:             "Starter Plan",
		RamMB:            1024,
		Vcpu:             2,
		StorageMB:        10240,
		MonthlyCostCents: 999,
	}

	resp := planToResponse(plan)

	assert.Equal(t, id.String(), resp.ID)
	assert.Equal(t, "starter", resp.Tier)
	assert.Equal(t, "Starter Plan", resp.Name)
	assert.Equal(t, 1024, resp.RamMB)
	assert.Equal(t, 2, resp.Vcpu)
	assert.Equal(t, 10240, resp.StorageMB)
	assert.Equal(t, 999, resp.MonthlyCostCents)
	assert.Equal(t, "9.99", resp.MonthlyCostUSD)
}

func TestPlanToResponse_ZeroCost(t *testing.T) {
	resp := planToResponse(types.MachinePlan{
		ID:               uuid.New(),
		MonthlyCostCents: 0,
	})
	assert.Equal(t, "0.00", resp.MonthlyCostUSD)
}

// ---------- generateKeyPair ----------

func TestGenerateKeyPair(t *testing.T) {
	priv, pub, fp, err := generateKeyPair()
	require.NoError(t, err)
	assert.NotEmpty(t, priv)
	assert.NotEmpty(t, pub)
	assert.NotEmpty(t, fp)
	assert.Contains(t, priv, "RSA PRIVATE KEY")
	assert.Contains(t, pub, "ssh-rsa")
	assert.True(t, len(fp) >= 7 && fp[:7] == "SHA256:", "fingerprint should start with SHA256:")
}

// ---------- ListPlans ----------

func TestBillingService_ListPlans_StorageError(t *testing.T) {
	repo := &mockBillingRepo{
		listActivePlansFn: func() ([]types.MachinePlan, error) { return nil, fmt.Errorf("db error") },
	}
	svc := NewBillingServiceWith(repo)
	_, err := svc.ListPlans()
	require.Error(t, err)
}

func TestBillingService_ListPlans_Success(t *testing.T) {
	plans := []types.MachinePlan{
		{ID: uuid.New(), Tier: "starter", Name: "Starter", MonthlyCostCents: 999},
		{ID: uuid.New(), Tier: "pro", Name: "Pro", MonthlyCostCents: 2999},
	}
	repo := &mockBillingRepo{
		listActivePlansFn: func() ([]types.MachinePlan, error) { return plans, nil },
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.ListPlans()
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Len(t, resp.Data, 2)
}

func TestBillingService_ListPlans_Empty(t *testing.T) {
	repo := &mockBillingRepo{
		listActivePlansFn: func() ([]types.MachinePlan, error) { return []types.MachinePlan{}, nil },
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.ListPlans()
	require.NoError(t, err)
	assert.Len(t, resp.Data, 0)
}

// ---------- SelectPlan ----------

func TestBillingService_SelectPlan_GetPlanError(t *testing.T) {
	repo := &mockBillingRepo{
		getPlanByTierFn: func(_ string) (*types.MachinePlan, error) { return nil, fmt.Errorf("db error") },
	}
	svc := NewBillingServiceWith(repo)
	_, err := svc.SelectPlan(context.Background(), uuid.New(), "starter")
	require.Error(t, err)
}

func TestBillingService_SelectPlan_PlanNotFound(t *testing.T) {
	repo := &mockBillingRepo{
		getPlanByTierFn: func(_ string) (*types.MachinePlan, error) { return nil, nil },
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.SelectPlan(context.Background(), uuid.New(), "nonexistent")
	require.NoError(t, err)
	assert.Equal(t, "error", resp.Status)
}

func TestBillingService_SelectPlan_GetBalanceError(t *testing.T) {
	plan := &types.MachinePlan{ID: uuid.New(), Tier: "starter", MonthlyCostCents: 999}
	repo := &mockBillingRepo{
		getPlanByTierFn:    func(_ string) (*types.MachinePlan, error) { return plan, nil },
		getWalletBalanceFn: func(_ uuid.UUID) (int, error) { return 0, fmt.Errorf("db error") },
	}
	svc := NewBillingServiceWith(repo)
	_, err := svc.SelectPlan(context.Background(), uuid.New(), "starter")
	require.Error(t, err)
}

func TestBillingService_SelectPlan_InsufficientBalance(t *testing.T) {
	plan := &types.MachinePlan{ID: uuid.New(), Tier: "starter", MonthlyCostCents: 999}
	repo := &mockBillingRepo{
		getPlanByTierFn:    func(_ string) (*types.MachinePlan, error) { return plan, nil },
		getWalletBalanceFn: func(_ uuid.UUID) (int, error) { return 100, nil },
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.SelectPlan(context.Background(), uuid.New(), "starter")
	require.NoError(t, err)
	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "insufficient_balance", resp.Error)
}

func TestBillingService_SelectPlan_DebitError(t *testing.T) {
	plan := &types.MachinePlan{ID: uuid.New(), Tier: "starter", MonthlyCostCents: 999}
	repo := &mockBillingRepo{
		getPlanByTierFn:    func(_ string) (*types.MachinePlan, error) { return plan, nil },
		getWalletBalanceFn: func(_ uuid.UUID) (int, error) { return 9999, nil },
		debitWalletFn:      func(_ uuid.UUID, _ int, _, _ string) (bool, error) { return false, fmt.Errorf("debit failed") },
	}
	svc := NewBillingServiceWith(repo)
	_, err := svc.SelectPlan(context.Background(), uuid.New(), "starter")
	require.Error(t, err)
}

func TestBillingService_SelectPlan_DebitReturnsFalse(t *testing.T) {
	plan := &types.MachinePlan{ID: uuid.New(), Tier: "starter", MonthlyCostCents: 999}
	repo := &mockBillingRepo{
		getPlanByTierFn:    func(_ string) (*types.MachinePlan, error) { return plan, nil },
		getWalletBalanceFn: func(_ uuid.UUID) (int, error) { return 9999, nil },
		debitWalletFn:      func(_ uuid.UUID, _ int, _, _ string) (bool, error) { return false, nil },
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.SelectPlan(context.Background(), uuid.New(), "starter")
	require.NoError(t, err)
	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "debit_failed", resp.Error)
}

func TestBillingService_SelectPlan_UpsertError(t *testing.T) {
	plan := &types.MachinePlan{ID: uuid.New(), Tier: "starter", MonthlyCostCents: 999}
	repo := &mockBillingRepo{
		getPlanByTierFn:    func(_ string) (*types.MachinePlan, error) { return plan, nil },
		getWalletBalanceFn: func(_ uuid.UUID) (int, error) { return 9999, nil },
		debitWalletFn:      func(_ uuid.UUID, _ int, _, _ string) (bool, error) { return true, nil },
		upsertBillingFn:    func(_ uuid.UUID, _ uuid.UUID, _, _ time.Time) error { return fmt.Errorf("upsert failed") },
	}
	svc := NewBillingServiceWith(repo)
	_, err := svc.SelectPlan(context.Background(), uuid.New(), "starter")
	require.Error(t, err)
}

func TestBillingService_SelectPlan_Success(t *testing.T) {
	plan := &types.MachinePlan{ID: uuid.New(), Tier: "starter", Name: "Starter", MonthlyCostCents: 999}
	repo := &mockBillingRepo{
		getPlanByTierFn:    func(_ string) (*types.MachinePlan, error) { return plan, nil },
		getWalletBalanceFn: func(_ uuid.UUID) (int, error) { return 9999, nil },
		debitWalletFn:      func(_ uuid.UUID, _ int, _, _ string) (bool, error) { return true, nil },
		upsertBillingFn:    func(_ uuid.UUID, _ uuid.UUID, _, _ time.Time) error { return nil },
		getProvisionInfoFn: func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (*storage.ProvisionInfo, error) {
			return nil, nil
		},
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.SelectPlan(context.Background(), uuid.New(), "starter")
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
}

func TestBillingService_SelectPlan_SuspendedReactivates(t *testing.T) {
	sshKeyID := uuid.New()
	plan := &types.MachinePlan{ID: uuid.New(), Tier: "starter", Name: "Starter", MonthlyCostCents: 999}
	existing := &types.OrgMachineBilling{
		MachinePlanID: plan.ID,
		Status:        types.MachineBillingStatusSuspended,
		SSHKeyID:      &sshKeyID,
	}
	reactivated := false
	repo := &mockBillingRepo{
		getPlanByTierFn:    func(_ string) (*types.MachinePlan, error) { return plan, nil },
		getWalletBalanceFn: func(_ uuid.UUID) (int, error) { return 9999, nil },
		debitWalletFn:      func(_ uuid.UUID, _ int, _, _ string) (bool, error) { return true, nil },
		upsertBillingFn:    func(_ uuid.UUID, _ uuid.UUID, _, _ time.Time) error { return nil },
		getBillingByOrgIDFn: func(_ uuid.UUID) (*types.OrgMachineBilling, error) {
			return existing, nil
		},
		reactivateSSHKeyFn: func(_ context.Context, _ uuid.UUID) error {
			reactivated = true
			return nil
		},
		getProvisionInfoFn: func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (*storage.ProvisionInfo, error) {
			return nil, nil
		},
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.SelectPlan(context.Background(), uuid.New(), "starter")
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.True(t, reactivated)
}

func TestBillingService_SelectPlan_EnqueueResourceUpgrade(t *testing.T) {
	plan := &types.MachinePlan{ID: uuid.New(), Tier: "starter", Name: "Starter", MonthlyCostCents: 999, Vcpu: 2, RamMB: 1024}
	repo := &mockBillingRepo{
		getPlanByTierFn:    func(_ string) (*types.MachinePlan, error) { return plan, nil },
		getWalletBalanceFn: func(_ uuid.UUID) (int, error) { return 9999, nil },
		debitWalletFn:      func(_ uuid.UUID, _ int, _, _ string) (bool, error) { return true, nil },
		upsertBillingFn:    func(_ uuid.UUID, _ uuid.UUID, _, _ time.Time) error { return nil },
		getProvisionInfoFn: func(_ context.Context, _ uuid.UUID, _ *uuid.UUID) (*storage.ProvisionInfo, error) {
			return &storage.ProvisionInfo{
				UserID:        uuid.New(),
				ContainerName: "container-1",
				ServerID:      "srv-1",
			}, nil
		},
	}
	svc := NewBillingServiceWith(repo)
	// This will call EnqueueResourceUpdateTask - it may fail in test env but should not return error
	resp, err := svc.SelectPlan(context.Background(), uuid.New(), "starter")
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
}

// ---------- IsServerUserOwned ----------

func TestBillingService_IsServerUserOwned_Error(t *testing.T) {
	repo := &mockBillingRepo{
		isServerUserOwnedFn: func(_ uuid.UUID, _ uuid.UUID) (bool, error) { return false, fmt.Errorf("db error") },
	}
	svc := NewBillingServiceWith(repo)
	_, err := svc.IsServerUserOwned(uuid.New(), uuid.New())
	require.Error(t, err)
}

func TestBillingService_IsServerUserOwned_Success(t *testing.T) {
	repo := &mockBillingRepo{
		isServerUserOwnedFn: func(_ uuid.UUID, _ uuid.UUID) (bool, error) { return true, nil },
	}
	svc := NewBillingServiceWith(repo)
	owned, err := svc.IsServerUserOwned(uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.True(t, owned)
}

// ---------- GetBillingStatus ----------

func TestBillingService_GetBillingStatus_StorageError(t *testing.T) {
	repo := &mockBillingRepo{
		getBillingByOrgIDFn: func(_ uuid.UUID) (*types.OrgMachineBilling, error) { return nil, fmt.Errorf("db error") },
	}
	svc := NewBillingServiceWith(repo)
	_, err := svc.GetBillingStatus(uuid.New())
	require.Error(t, err)
}

func TestBillingService_GetBillingStatus_NoBilling_NoSSH(t *testing.T) {
	repo := &mockBillingRepo{
		getBillingByOrgIDFn: func(_ uuid.UUID) (*types.OrgMachineBilling, error) { return nil, nil },
		hasActiveSSHKeyFn:   func(_ uuid.UUID) (bool, error) { return false, nil },
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.GetBillingStatus(uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.False(t, resp.Data.HasMachine)
}

func TestBillingService_GetBillingStatus_NoBilling_HasSSH(t *testing.T) {
	repo := &mockBillingRepo{
		getBillingByOrgIDFn: func(_ uuid.UUID) (*types.OrgMachineBilling, error) { return nil, nil },
		hasActiveSSHKeyFn:   func(_ uuid.UUID) (bool, error) { return true, nil },
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.GetBillingStatus(uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.True(t, resp.Data.HasMachine)
	assert.Equal(t, "unbilled", resp.Data.BillingStatus)
}

func TestBillingService_GetBillingStatus_HasSSHError(t *testing.T) {
	repo := &mockBillingRepo{
		getBillingByOrgIDFn: func(_ uuid.UUID) (*types.OrgMachineBilling, error) { return nil, nil },
		hasActiveSSHKeyFn:   func(_ uuid.UUID) (bool, error) { return false, fmt.Errorf("db error") },
	}
	svc := NewBillingServiceWith(repo)
	_, err := svc.GetBillingStatus(uuid.New())
	require.Error(t, err)
}

func TestBillingService_GetBillingStatus_WithBilling_Active(t *testing.T) {
	planID := uuid.New()
	periodEnd := time.Now().Add(30 * 24 * time.Hour)
	billing := &types.OrgMachineBilling{
		MachinePlanID:    planID,
		Status:           types.MachineBillingStatusActive,
		CurrentPeriodEnd: periodEnd,
	}
	plan := &types.MachinePlan{ID: planID, Tier: "starter", Name: "Starter", MonthlyCostCents: 999}
	repo := &mockBillingRepo{
		getBillingByOrgIDFn: func(_ uuid.UUID) (*types.OrgMachineBilling, error) { return billing, nil },
		getPlanByIDFn:       func(_ uuid.UUID) (*types.MachinePlan, error) { return plan, nil },
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.GetBillingStatus(uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.True(t, resp.Data.HasMachine)
	assert.Equal(t, "active", resp.Data.BillingStatus)
	assert.Equal(t, "starter", resp.Data.PlanTier)
}

func TestBillingService_GetBillingStatus_WithBilling_PlanError(t *testing.T) {
	planID := uuid.New()
	billing := &types.OrgMachineBilling{
		MachinePlanID:    planID,
		Status:           types.MachineBillingStatusActive,
		CurrentPeriodEnd: time.Now().Add(30 * 24 * time.Hour),
	}
	repo := &mockBillingRepo{
		getBillingByOrgIDFn: func(_ uuid.UUID) (*types.OrgMachineBilling, error) { return billing, nil },
		getPlanByIDFn:       func(_ uuid.UUID) (*types.MachinePlan, error) { return nil, fmt.Errorf("plan db error") },
	}
	svc := NewBillingServiceWith(repo)
	_, err := svc.GetBillingStatus(uuid.New())
	require.Error(t, err)
}

func TestBillingService_GetBillingStatus_WithBilling_NilPlan(t *testing.T) {
	planID := uuid.New()
	billing := &types.OrgMachineBilling{
		MachinePlanID:    planID,
		Status:           types.MachineBillingStatusActive,
		CurrentPeriodEnd: time.Now().Add(30 * 24 * time.Hour),
	}
	repo := &mockBillingRepo{
		getBillingByOrgIDFn: func(_ uuid.UUID) (*types.OrgMachineBilling, error) { return billing, nil },
		getPlanByIDFn:       func(_ uuid.UUID) (*types.MachinePlan, error) { return nil, nil },
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.GetBillingStatus(uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
}

func TestBillingService_GetBillingStatus_GracePeriod(t *testing.T) {
	planID := uuid.New()
	grace := time.Now().Add(3 * 24 * time.Hour)
	billing := &types.OrgMachineBilling{
		MachinePlanID:    planID,
		Status:           types.MachineBillingStatusGracePeriod,
		CurrentPeriodEnd: time.Now().Add(30 * 24 * time.Hour),
		GraceDeadline:    &grace,
	}
	plan := &types.MachinePlan{ID: planID, Tier: "starter", Name: "Starter", MonthlyCostCents: 999}
	repo := &mockBillingRepo{
		getBillingByOrgIDFn: func(_ uuid.UUID) (*types.OrgMachineBilling, error) { return billing, nil },
		getPlanByIDFn:       func(_ uuid.UUID) (*types.MachinePlan, error) { return plan, nil },
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.GetBillingStatus(uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "grace_period", resp.Data.BillingStatus)
	assert.NotNil(t, resp.Data.DaysRemaining)
	assert.NotEmpty(t, resp.Data.Message)
}

func TestBillingService_GetBillingStatus_GracePeriod_PastDeadline(t *testing.T) {
	planID := uuid.New()
	grace := time.Now().Add(-24 * time.Hour) // past deadline
	billing := &types.OrgMachineBilling{
		MachinePlanID:    planID,
		Status:           types.MachineBillingStatusGracePeriod,
		CurrentPeriodEnd: time.Now().Add(30 * 24 * time.Hour),
		GraceDeadline:    &grace,
	}
	plan := &types.MachinePlan{ID: planID, Tier: "starter", MonthlyCostCents: 999}
	repo := &mockBillingRepo{
		getBillingByOrgIDFn: func(_ uuid.UUID) (*types.OrgMachineBilling, error) { return billing, nil },
		getPlanByIDFn:       func(_ uuid.UUID) (*types.MachinePlan, error) { return plan, nil },
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.GetBillingStatus(uuid.New())
	require.NoError(t, err)
	assert.Equal(t, 0, *resp.Data.DaysRemaining)
}

func TestBillingService_GetBillingStatus_Suspended(t *testing.T) {
	planID := uuid.New()
	billing := &types.OrgMachineBilling{
		MachinePlanID:    planID,
		Status:           types.MachineBillingStatusSuspended,
		CurrentPeriodEnd: time.Now().Add(30 * 24 * time.Hour),
	}
	plan := &types.MachinePlan{ID: planID, Tier: "starter", MonthlyCostCents: 999}
	repo := &mockBillingRepo{
		getBillingByOrgIDFn: func(_ uuid.UUID) (*types.OrgMachineBilling, error) { return billing, nil },
		getPlanByIDFn:       func(_ uuid.UUID) (*types.MachinePlan, error) { return plan, nil },
	}
	svc := NewBillingServiceWith(repo)
	resp, err := svc.GetBillingStatus(uuid.New())
	require.NoError(t, err)
	assert.Equal(t, "suspended", resp.Data.BillingStatus)
	assert.NotEmpty(t, resp.Data.Message)
}

func TestBillingService_CheckUnpaidTrial_Error(t *testing.T) {
	repo := &mockBillingRepo{
		hasTrialWithoutActiveBillingFn: func(_ uuid.UUID) (bool, error) { return false, fmt.Errorf("db error") },
	}
	svc := NewBillingServiceWith(repo)
	// checkUnpaidTrial returns false on error (non-fatal)
	result := svc.checkUnpaidTrial(uuid.New())
	assert.False(t, result)
}

func TestBillingService_CheckUnpaidTrial_True(t *testing.T) {
	repo := &mockBillingRepo{
		hasTrialWithoutActiveBillingFn: func(_ uuid.UUID) (bool, error) { return true, nil },
	}
	svc := NewBillingServiceWith(repo)
	result := svc.checkUnpaidTrial(uuid.New())
	assert.True(t, result)
}
