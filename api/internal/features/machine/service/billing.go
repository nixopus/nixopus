package service

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/queue"
)

type BillingService struct {
	storage *storage.BillingStorage
}

func NewBillingService(s *storage.BillingStorage) *BillingService {
	return &BillingService{storage: s}
}

func planToResponse(p types.MachinePlan) types.MachinePlanResponse {
	return types.MachinePlanResponse{
		ID:               p.ID.String(),
		Tier:             p.Tier,
		Name:             p.Name,
		RamMB:            p.RamMB,
		Vcpu:             p.Vcpu,
		StorageMB:        p.StorageMB,
		MonthlyCostCents: p.MonthlyCostCents,
		MonthlyCostUSD:   fmt.Sprintf("%.2f", float64(p.MonthlyCostCents)/100),
	}
}

func (s *BillingService) ListPlans() (*types.ListPlansResponse, error) {
	plans, err := s.storage.ListActivePlans()
	if err != nil {
		return nil, err
	}

	data := make([]types.MachinePlanResponse, len(plans))
	for i, p := range plans {
		data[i] = planToResponse(p)
	}

	return &types.ListPlansResponse{
		Status: "success",
		Data:   data,
	}, nil
}

func (s *BillingService) SelectPlan(ctx context.Context, orgID uuid.UUID, planTier string) (*types.SelectPlanResponse, error) {
	plan, err := s.storage.GetPlanByTier(planTier)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		return &types.SelectPlanResponse{
			Status:  "error",
			Message: fmt.Sprintf("Plan tier %q not found or inactive.", planTier),
			Error:   fmt.Sprintf("Plan tier %q not found or inactive.", planTier),
		}, nil
	}

	balance, err := s.storage.GetWalletBalance(orgID)
	if err != nil {
		return nil, err
	}

	if balance < plan.MonthlyCostCents {
		return &types.SelectPlanResponse{
			Status:  "error",
			Message: fmt.Sprintf("Insufficient wallet balance. Need $%.2f but wallet has $%.2f. Top up your wallet first.", float64(plan.MonthlyCostCents)/100, float64(balance)/100),
			Error:   "insufficient_balance",
		}, nil
	}

	existing, _ := s.storage.GetBillingByOrgID(orgID)
	wasSuspended := existing != nil && existing.Status == types.MachineBillingStatusSuspended

	now := time.Now()
	periodEnd := now.AddDate(0, 1, 0)
	refID := fmt.Sprintf("machine_select_%s_%s_%s", orgID.String(), plan.Tier, now.Format("2006-01-02"))

	debited, err := s.storage.DebitWallet(orgID, plan.MonthlyCostCents, fmt.Sprintf("Machine plan selected: %s", plan.Name), refID)
	if err != nil {
		return nil, err
	}
	if !debited {
		return &types.SelectPlanResponse{
			Status:  "error",
			Message: "Wallet debit failed. Balance may have changed.",
			Error:   "debit_failed",
		}, nil
	}

	err = s.storage.UpsertBilling(orgID, plan.ID, now, periodEnd)
	if err != nil {
		return nil, err
	}

	var billingSSHKeyID *uuid.UUID
	if existing != nil {
		billingSSHKeyID = existing.SSHKeyID
	}

	if wasSuspended && existing.SSHKeyID != nil {
		_ = s.storage.ReactivateSSHKey(ctx, *existing.SSHKeyID)
	}

	s.enqueueResourceUpgrade(ctx, orgID, plan, billingSSHKeyID)

	newBalance, _ := s.storage.GetWalletBalance(orgID)

	pr := planToResponse(*plan)
	return &types.SelectPlanResponse{
		Status:            "success",
		Message:           fmt.Sprintf("Plan %s selected. $%.2f charged for the first month.", plan.Name, float64(plan.MonthlyCostCents)/100),
		Plan:              &pr,
		ChargedCents:      plan.MonthlyCostCents,
		BalanceAfterCents: newBalance,
		PeriodEnd:         periodEnd.Format(time.RFC3339),
	}, nil
}

func (s *BillingService) enqueueResourceUpgrade(ctx context.Context, orgID uuid.UUID, plan *types.MachinePlan, sshKeyID *uuid.UUID) {
	info, err := s.storage.GetProvisionInfo(ctx, orgID, sshKeyID)
	if err != nil || info == nil || info.ContainerName == "" {
		return
	}

	_ = queue.EnqueueResourceUpdateTask(ctx, queue.ResourceUpdatePayload{
		VMName:    info.ContainerName,
		VcpuCount: plan.Vcpu,
		MemoryMB:  plan.RamMB,
		UserID:    info.UserID.String(),
		OrgID:     orgID.String(),
		ServerID:  info.ServerID,
	})
}

func (s *BillingService) GetBillingStatus(orgID uuid.UUID) (*types.MachineBillingResponse, error) {
	billing, err := s.storage.GetBillingByOrgID(orgID)
	if err != nil {
		return nil, err
	}

	hasUnpaidTrial := s.checkUnpaidTrial(orgID)

	if billing != nil {
		plan, err := s.storage.GetPlanByID(billing.MachinePlanID)
		if err != nil {
			return nil, err
		}

		data := &types.MachineBillingStatusData{
			HasMachine:     true,
			BillingStatus:  string(billing.Status),
			PeriodEnd:      billing.CurrentPeriodEnd.Format(time.RFC3339),
			HasUnpaidTrial: hasUnpaidTrial,
		}

		if plan != nil {
			data.PlanTier = plan.Tier
			data.PlanName = plan.Name
			data.MonthlyCostCents = plan.MonthlyCostCents
			data.MonthlyCostUSD = fmt.Sprintf("%.2f", float64(plan.MonthlyCostCents)/100)
		}

		if billing.Status == types.MachineBillingStatusGracePeriod && billing.GraceDeadline != nil {
			data.GraceDeadline = billing.GraceDeadline.Format(time.RFC3339)
			days := int(math.Ceil(time.Until(*billing.GraceDeadline).Hours() / 24))
			if days < 0 {
				days = 0
			}
			data.DaysRemaining = &days
			data.Message = fmt.Sprintf("Your server will be reset in %d day(s). Top up your wallet to keep your server.", days)
		}

		if billing.Status == types.MachineBillingStatusSuspended {
			data.Message = "Your server was reset due to insufficient wallet balance. Top up your wallet and select a machine plan to restore service."
		}

		return &types.MachineBillingResponse{Status: "success", Data: data}, nil
	}

	hasSSH, err := s.storage.HasActiveSSHKey(orgID)
	if err != nil {
		return nil, err
	}

	if hasSSH {
		return &types.MachineBillingResponse{
			Status: "success",
			Data: &types.MachineBillingStatusData{
				HasMachine:     true,
				BillingStatus:  "unbilled",
				Message:        "Your machine does not have a billing plan configured.",
				HasUnpaidTrial: hasUnpaidTrial,
			},
		}, nil
	}

	return &types.MachineBillingResponse{
		Status: "success",
		Data:   &types.MachineBillingStatusData{HasMachine: false, HasUnpaidTrial: hasUnpaidTrial},
	}, nil
}

func (s *BillingService) checkUnpaidTrial(orgID uuid.UUID) bool {
	hasTrialWithoutBilling, err := s.storage.HasTrialWithoutActiveBilling(orgID)
	if err != nil {
		return false
	}
	return hasTrialWithoutBilling
}
