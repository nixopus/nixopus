package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	machine_storage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/queue"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/robfig/cron/v3"
	"github.com/uptrace/bun"
)

const (
	backupScheduleCheck = "0 * * * *"

	dailyMinGap  = 20 * time.Hour
	weeklyMinGap = 6 * 24 * time.Hour
)

type BackupScheduler struct {
	cron         *cron.Cron
	billingStore *machine_storage.BillingStorage
	backupStore  *machine_storage.BackupStorage
	db           *bun.DB
	logger       logger.Logger
	ctx          context.Context
}

func NewBackupScheduler(db *bun.DB, ctx context.Context, l logger.Logger) *BackupScheduler {
	return &BackupScheduler{
		cron:         cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger))),
		billingStore: machine_storage.NewBillingStorage(db, ctx),
		backupStore:  machine_storage.NewBackupStorage(db, ctx),
		db:           db,
		logger:       l,
		ctx:          ctx,
	}
}

func (b *BackupScheduler) Start() {
	_, err := b.cron.AddFunc(backupScheduleCheck, b.run)
	if err != nil {
		b.logger.Log(logger.Error, fmt.Sprintf("backup scheduler: failed to register cron: %v", err), "")
		return
	}
	b.cron.Start()
	b.logger.Log(logger.Info, fmt.Sprintf("backup scheduler started with schedule: %s", backupScheduleCheck), "")
}

func (b *BackupScheduler) Stop() {
	b.cron.Stop()
}

func (b *BackupScheduler) run() {
	now := time.Now().UTC()
	currentHour := now.Hour()
	currentDay := int(now.Weekday())

	var orgSettings []*types.OrganizationSettings
	err := b.db.NewSelect().
		Model(&orgSettings).
		Scan(b.ctx)
	if err != nil {
		b.logger.Log(logger.Error, fmt.Sprintf("backup scheduler: failed to load org settings: %v", err), "")
		return
	}

	enqueued, skipped := 0, 0
	for _, org := range orgSettings {
		s := org.Settings
		if !isBackupEnabled(s) {
			continue
		}

		configuredHour := 2
		if s.BackupScheduleHourUTC != nil {
			configuredHour = *s.BackupScheduleHourUTC
		}
		if currentHour != configuredHour {
			continue
		}

		freq := "daily"
		if s.BackupScheduleFrequency != nil {
			freq = *s.BackupScheduleFrequency
		}
		if freq == "weekly" {
			configuredDay := 0
			if s.BackupScheduleDayOfWeek != nil {
				configuredDay = *s.BackupScheduleDayOfWeek
			}
			if currentDay != configuredDay {
				continue
			}
		}

		orgID := org.OrganizationID
		if err := b.enqueueIfDue(orgID, freq, now); err != nil {
			b.logger.Log(logger.Error, fmt.Sprintf("backup scheduler: org %s: %v", orgID, err), "")
			skipped++
		} else {
			enqueued++
		}
	}

	if enqueued > 0 || skipped > 0 {
		b.logger.Log(logger.Info, fmt.Sprintf("backup scheduler: enqueued %d, skipped %d", enqueued, skipped), "")
	}
}

func (b *BackupScheduler) enqueueIfDue(orgID uuid.UUID, freq string, now time.Time) error {
	hasRunning, err := b.backupStore.HasInProgressBackup(b.ctx, orgID)
	if err != nil {
		return fmt.Errorf("check in-progress: %w", err)
	}
	if hasRunning {
		return fmt.Errorf("backup already in progress")
	}

	latest, err := b.backupStore.GetLatestCompletedBackup(b.ctx, orgID)
	if err != nil {
		return fmt.Errorf("check latest backup: %w", err)
	}
	if latest != nil && latest.CompletedAt != nil {
		minGap := dailyMinGap
		if freq == "weekly" {
			minGap = weeklyMinGap
		}
		if now.Sub(*latest.CompletedAt) < minGap {
			return fmt.Errorf("last backup too recent (%s ago)", now.Sub(*latest.CompletedAt).Round(time.Minute))
		}
	}

	info, err := b.billingStore.GetProvisionInfo(b.ctx, orgID, nil)
	if err != nil {
		return fmt.Errorf("resolve machine: %w", err)
	}
	if info == nil || info.ContainerName == "" {
		return fmt.Errorf("no provisioned machine")
	}

	payload := queue.MachineBackupPayload{
		MachineName: info.ContainerName,
		UserID:      info.UserID.String(),
		OrgID:       orgID.String(),
		ServerID:    info.ServerID,
		Trigger:     "scheduled",
	}

	requestID, err := queue.EnqueueMachineBackup(b.ctx, payload)
	if err != nil {
		return fmt.Errorf("enqueue: %w", err)
	}

	b.logger.Log(logger.Info, fmt.Sprintf("backup scheduler: enqueued backup for org %s machine %s (request %s)", orgID, info.ContainerName, requestID), "")
	return nil
}

func isBackupEnabled(s types.OrganizationSettingsData) bool {
	return s.BackupScheduleEnabled != nil && *s.BackupScheduleEnabled
}
