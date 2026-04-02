package scheduler

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	trail_storage "github.com/nixopus/nixopus/api/internal/features/trail/storage"
	"github.com/nixopus/nixopus/api/internal/queue"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/robfig/cron/v3"
	"github.com/uptrace/bun"
)

const trialExpirySchedule = "0 * * * *"

type TrialExpiryScheduler struct {
	cron       *cron.Cron
	storage    *trail_storage.TrailStorage
	logger     logger.Logger
	ctx        context.Context
	trialDays  int
	notifierMu sync.RWMutex
	notifier   shared_types.Notifier
}

func NewTrialExpiryScheduler(db *bun.DB, ctx context.Context, l logger.Logger, trialDays int) *TrialExpiryScheduler {
	if trialDays <= 0 {
		trialDays = 7
	}
	return &TrialExpiryScheduler{
		cron:      cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger))),
		storage:   trail_storage.NewTrailStorage(db, ctx),
		logger:    l,
		ctx:       ctx,
		trialDays: trialDays,
	}
}

func (t *TrialExpiryScheduler) SetNotifier(n shared_types.Notifier) {
	t.notifierMu.Lock()
	defer t.notifierMu.Unlock()
	t.notifier = n
}

func (t *TrialExpiryScheduler) getNotifier() shared_types.Notifier {
	t.notifierMu.RLock()
	defer t.notifierMu.RUnlock()
	return t.notifier
}

func (t *TrialExpiryScheduler) Start() {
	_, err := t.cron.AddFunc(trialExpirySchedule, t.run)
	if err != nil {
		t.logger.Log(logger.Error, fmt.Sprintf("trial expiry scheduler: failed to register cron: %v", err), "")
		return
	}
	t.cron.Start()
	t.logger.Log(logger.Info, fmt.Sprintf("trial expiry scheduler started with schedule: %s (trial_period_days=%d)", trialExpirySchedule, t.trialDays), "")
}

func (t *TrialExpiryScheduler) Stop() {
	t.cron.Stop()
}

func (t *TrialExpiryScheduler) run() {
	users, err := t.storage.GetExpiredTrialUsers(t.ctx, t.trialDays)
	if err != nil {
		t.logger.Log(logger.Error, fmt.Sprintf("trial expiry: query failed: %v", err), "")
		return
	}

	if len(users) == 0 {
		return
	}

	t.logger.Log(logger.Info, fmt.Sprintf("trial expiry: found %d expired trial user(s)", len(users)), "")

	processed, skipped, failed := 0, 0, 0
	for _, user := range users {
		if user.LXDContainerName == nil || *user.LXDContainerName == "" {
			t.logger.Log(logger.Warning, fmt.Sprintf("trial expiry: skipping user %s — no container name", user.UserID), user.UserID.String())
			skipped++
			continue
		}

		hasBilling, err := t.storage.HasMachineBilling(t.ctx, user.OrganizationID)
		if err != nil {
			t.logger.Log(logger.Error, fmt.Sprintf("trial expiry: billing re-check failed for user %s: %v", user.UserID, err), user.UserID.String())
			failed++
			continue
		}
		if hasBilling {
			t.logger.Log(logger.Info, fmt.Sprintf("trial expiry: skipping user %s — billing record appeared", user.UserID), user.UserID.String())
			skipped++
			continue
		}

		serverID := ""
		if user.ServerID != nil {
			serverID = user.ServerID.String()
		}

		if err := queue.EnqueueVMDeleteTask(t.ctx, queue.VMDeletePayload{
			VMName:   *user.LXDContainerName,
			UserID:   user.UserID.String(),
			OrgID:    user.OrganizationID.String(),
			ServerID: serverID,
		}); err != nil {
			t.logger.Log(logger.Error, fmt.Sprintf("trial expiry: failed to enqueue delete for user %s: %v", user.UserID, err), user.UserID.String())
			failed++
			continue
		}

		if err := t.storage.DeleteProvisionAndResetStatus(t.ctx, user.ProvisionID, user.UserID); err != nil {
			t.logger.Log(logger.Error, fmt.Sprintf("trial expiry: failed to reset provision for user %s: %v", user.UserID, err), user.UserID.String())
			failed++
			continue
		}

		if notifier := t.getNotifier(); notifier != nil {
			if err := notifier.Emit(shared_types.NotificationEvent{
				Type:           shared_types.EventTrialExpired,
				UserID:         user.UserID.String(),
				OrganizationID: user.OrganizationID.String(),
				Data: map[string]interface{}{
					"name":       user.Name,
					"email":      user.Email,
					"trial_days": fmt.Sprintf("%d", t.trialDays),
				},
			}); err != nil {
				t.logger.Log(logger.Warning, fmt.Sprintf("trial expiry: notification failed for user %s: %v", user.UserID, err), user.UserID.String())
			}
		} else {
			log.Printf("[trial-expiry] notifier not available, skipping notification for user %s", user.UserID)
		}

		processed++
		t.logger.Log(logger.Info, fmt.Sprintf("trial expiry: processed user %s (%s)", user.UserID, user.Email), user.UserID.String())
	}

	t.logger.Log(logger.Info, fmt.Sprintf("trial expiry: done — processed=%d skipped=%d failed=%d", processed, skipped, failed), "")
}
