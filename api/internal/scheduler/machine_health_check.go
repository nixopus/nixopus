package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	machine_storage "github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/queue"
	"github.com/nixopus/nixopus/api/internal/types"
	"github.com/robfig/cron/v3"
	"github.com/uptrace/bun"
)

const machineHealthCheckSchedule = "*/30 * * * *"

type MachineHealthCheckScheduler struct {
	cron    *cron.Cron
	storage *machine_storage.RegistrationStorage
	db      *bun.DB
	logger  logger.Logger
	ctx     context.Context
}

func NewMachineHealthCheckScheduler(db *bun.DB, ctx context.Context, l logger.Logger) *MachineHealthCheckScheduler {
	return &MachineHealthCheckScheduler{
		cron:    cron.New(cron.WithChain(cron.Recover(cron.DefaultLogger))),
		storage: machine_storage.NewRegistrationStorage(db, ctx),
		db:      db,
		logger:  l,
		ctx:     ctx,
	}
}

func (s *MachineHealthCheckScheduler) Start() {
	_, err := s.cron.AddFunc(machineHealthCheckSchedule, s.run)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("machine health check: failed to register cron: %v", err), "")
		return
	}
	s.cron.Start()
	s.logger.Log(logger.Info, fmt.Sprintf("machine health check scheduler started with schedule: %s", machineHealthCheckSchedule), "")
}

func (s *MachineHealthCheckScheduler) Stop() {
	s.cron.Stop()
}

func (s *MachineHealthCheckScheduler) run() {
	var orgSettings []*types.OrganizationSettings
	err := s.db.NewSelect().
		Model(&orgSettings).
		Scan(s.ctx)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("machine health check: failed to load orgs: %v", err), "")
		return
	}

	serverID, _ := s.storage.GetAnyActiveInfraServerID()

	recentCutoff := time.Now().Add(-25 * time.Minute)
	enqueued := 0

	for _, org := range orgSettings {
		machines, err := s.storage.GetActiveUserOwnedMachines(org.OrganizationID)
		if err != nil {
			s.logger.Log(logger.Error, fmt.Sprintf("machine health check: org %s: %v", org.OrganizationID, err), "")
			continue
		}

		for _, machine := range machines {
			if machine.LastUsedAt != nil && machine.LastUsedAt.After(recentCutoff) {
				continue
			}

			payload := queue.MachineVerifyPayload{
				MachineID: machine.ID.String(),
				OrgID:     org.OrganizationID.String(),
				ServerID:  serverID,
			}
			if err := queue.EnqueueMachineVerifyTask(s.ctx, payload); err != nil {
				s.logger.Log(logger.Error, fmt.Sprintf("machine health check: enqueue failed for %s: %v", machine.ID, err), "")
				continue
			}
			enqueued++
		}
	}

	if enqueued > 0 {
		s.logger.Log(logger.Info, fmt.Sprintf("machine health check: enqueued %d verification tasks", enqueued), "")
	}
}
