package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	machine_storage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/robfig/cron/v3"
	"github.com/uptrace/bun"
)

const staleMachineCleanupSchedule = "0 * * * *"

type StaleMachineCleanupScheduler struct {
	cron    *cron.Cron
	storage *machine_storage.RegistrationStorage
	db      *bun.DB
	logger  logger.Logger
	ctx     context.Context
}

func NewStaleMachineCleanupScheduler(db *bun.DB, ctx context.Context, l logger.Logger) *StaleMachineCleanupScheduler {
	return &StaleMachineCleanupScheduler{
		cron:    cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger))),
		storage: machine_storage.NewRegistrationStorage(db, ctx),
		db:      db,
		logger:  l,
		ctx:     ctx,
	}
}

func (s *StaleMachineCleanupScheduler) Start() {
	_, err := s.cron.AddFunc(staleMachineCleanupSchedule, s.run)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("stale machine cleanup: failed to register cron: %v", err), "")
		return
	}
	s.cron.Start()
	s.logger.Log(logger.Info, fmt.Sprintf("stale machine cleanup scheduler started with schedule: %s", staleMachineCleanupSchedule), "")
}

func (s *StaleMachineCleanupScheduler) Stop() {
	s.cron.Stop()
}

func (s *StaleMachineCleanupScheduler) run() {
	var orgSettings []*types.OrganizationSettings
	err := s.db.NewSelect().
		Model(&orgSettings).
		Scan(s.ctx)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("stale machine cleanup: failed to load orgs: %v", err), "")
		return
	}

	cutoff := time.Now().Add(-24 * time.Hour)
	cleaned := 0

	for _, org := range orgSettings {
		ids, err := s.storage.GetStaleBYOSMachines(org.OrganizationID, cutoff)
		if err != nil {
			s.logger.Log(logger.Error, fmt.Sprintf("stale machine cleanup: org %s: %v", org.OrganizationID, err), "")
			continue
		}
		for _, id := range ids {
			if err := s.storage.SoftDeleteSSHKey(id); err != nil {
				s.logger.Log(logger.Error, fmt.Sprintf("stale machine cleanup: failed to delete %s: %v", id, err), "")
				continue
			}
			cleaned++
		}
	}

	if cleaned > 0 {
		s.logger.Log(logger.Info, fmt.Sprintf("stale machine cleanup: removed %d stale machines", cleaned), "")
	}
}
