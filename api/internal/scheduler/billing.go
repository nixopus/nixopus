package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	billing_storage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/queue"
	"github.com/robfig/cron/v3"
	"github.com/uptrace/bun"
)

const (
	gracePeriodDays    = 7
	billingJobSchedule = "0 * * * *"
)

type BillingScheduler struct {
	cron    *cron.Cron
	storage *billing_storage.BillingStorage
	db      *bun.DB
	logger  logger.Logger
	ctx     context.Context
}

func NewBillingScheduler(db *bun.DB, ctx context.Context, l logger.Logger) *BillingScheduler {
	return &BillingScheduler{
		cron:    cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger))),
		storage: billing_storage.NewBillingStorage(db, ctx),
		db:      db,
		logger:  l,
		ctx:     ctx,
	}
}

func (b *BillingScheduler) Start() {
	_, err := b.cron.AddFunc(billingJobSchedule, b.run)
	if err != nil {
		b.logger.Log(logger.Error, fmt.Sprintf("billing scheduler: failed to register cron: %v", err), "")
		return
	}
	b.cron.Start()
	b.logger.Log(logger.Info, fmt.Sprintf("billing scheduler started with schedule: %s", billingJobSchedule), "")
}

func (b *BillingScheduler) Stop() {
	b.cron.Stop()
}

func (b *BillingScheduler) run() {
	b.runSweep()
	b.runGraceCheck()
}

func (b *BillingScheduler) runSweep() {
	rows, err := b.storage.GetDueBillings(b.ctx)
	if err != nil {
		b.logger.Log(logger.Error, fmt.Sprintf("billing sweep: failed to get due billings: %v", err), "")
		return
	}

	if len(rows) == 0 {
		return
	}

	charged, grace := 0, 0
	for _, row := range rows {
		orgID := row.Billing.OrganizationID
		costCents := row.Plan.MonthlyCostCents
		periodEnd := row.Billing.CurrentPeriodEnd

		nextStart := periodEnd
		nextEnd := periodEnd.AddDate(0, 1, 0)
		refID := fmt.Sprintf("machine_billing_%s_%s", row.Billing.ID.String(), nextStart.Format("2006-01"))

		debited, err := b.storage.DebitWallet(orgID, costCents, fmt.Sprintf("Machine plan: %s", row.Plan.Name), refID)
		if err != nil {
			b.logger.Log(logger.Error, fmt.Sprintf("billing sweep: debit failed for org %s: %v", orgID, err), "")
			continue
		}

		if debited {
			if err := b.storage.UpdateBillingPeriod(b.ctx, row.Billing.ID, nextStart, nextEnd); err != nil {
				b.logger.Log(logger.Error, fmt.Sprintf("billing sweep: update period failed for org %s: %v", orgID, err), "")
			}
			charged++
			b.logger.Log(logger.Info, fmt.Sprintf("billing sweep: charged org %s %d cents for %s", orgID, costCents, row.Plan.Tier), "")
		} else {
			deadline := time.Now().AddDate(0, 0, gracePeriodDays)
			if err := b.storage.SetGracePeriod(b.ctx, row.Billing.ID, deadline); err != nil {
				b.logger.Log(logger.Error, fmt.Sprintf("billing sweep: set grace failed for org %s: %v", orgID, err), "")
			}
			grace++
			b.logger.Log(logger.Warning, fmt.Sprintf("billing sweep: insufficient funds for org %s, grace period until %s", orgID, deadline.Format(time.RFC3339)), "")
		}
	}

	b.logger.Log(logger.Info, fmt.Sprintf("billing sweep: processed %d, charged %d, grace %d", len(rows), charged, grace), "")
}

func (b *BillingScheduler) runGraceCheck() {
	rows, err := b.storage.GetGraceBillings(b.ctx)
	if err != nil {
		b.logger.Log(logger.Error, fmt.Sprintf("grace check: failed to get grace billings: %v", err), "")
		return
	}

	if len(rows) == 0 {
		return
	}

	recovered, suspended, waiting := 0, 0, 0
	for _, row := range rows {
		orgID := row.Billing.OrganizationID
		costCents := row.Plan.MonthlyCostCents
		periodEnd := row.Billing.CurrentPeriodEnd

		balance, _ := b.storage.GetWalletBalance(orgID)
		if balance >= costCents {
			nextStart := periodEnd
			nextEnd := periodEnd.AddDate(0, 1, 0)
			refID := fmt.Sprintf("machine_billing_%s_%s_recovered", row.Billing.ID.String(), nextStart.Format("2006-01"))

			debited, _ := b.storage.DebitWallet(orgID, costCents, fmt.Sprintf("Machine plan: %s (recovered)", row.Plan.Name), refID)
			if debited {
				b.storage.RecoverFromGrace(b.ctx, row.Billing.ID, nextStart, nextEnd)
				recovered++
				b.logger.Log(logger.Info, fmt.Sprintf("grace check: recovered org %s", orgID), "")
				continue
			}
		}

		now := time.Now()
		if row.Billing.GraceDeadline != nil && !row.Billing.GraceDeadline.After(now) {
			b.storage.SuspendBilling(b.ctx, row.Billing.ID)

			if row.Billing.SSHKeyID != nil {
				b.storage.DeactivateSSHKey(b.ctx, *row.Billing.SSHKeyID)
			}

			b.triggerServerReset(orgID, row.Billing.SSHKeyID)
			suspended++
			b.logger.Log(logger.Warning, fmt.Sprintf("grace check: suspended org %s, server reset triggered", orgID), "")
		} else {
			waiting++
		}
	}

	if recovered > 0 || suspended > 0 {
		b.logger.Log(logger.Info, fmt.Sprintf("grace check: processed %d, recovered %d, suspended %d, waiting %d", len(rows), recovered, suspended, waiting), "")
	}
}

func (b *BillingScheduler) triggerServerReset(orgID uuid.UUID, sshKeyID *uuid.UUID) {
	info, err := b.storage.GetProvisionInfo(b.ctx, orgID, sshKeyID)
	if err != nil || info == nil {
		b.logger.Log(logger.Error, fmt.Sprintf("server reset: no provision found for org %s", orgID), "")
		return
	}

	if info.ContainerName == "" {
		b.logger.Log(logger.Error, fmt.Sprintf("server reset: no container name for org %s", orgID), "")
		return
	}

	payload := queue.ResourceUpdatePayload{
		VMName:    info.ContainerName,
		VcpuCount: 0,
		MemoryMB:  0,
		UserID:    info.UserID.String(),
		OrgID:     orgID.String(),
		ServerID:  info.ServerID,
	}

	if err := queue.EnqueueResourceUpdateTask(b.ctx, payload); err != nil {
		b.logger.Log(logger.Error, fmt.Sprintf("server reset: failed to enqueue for org %s: %v", orgID, err), "")
		return
	}

	b.logger.Log(logger.Info, fmt.Sprintf("server reset: enqueued for org %s user %s", orgID, info.UserID), "")
}
