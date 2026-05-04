package storage

import (
	"context"
	"fmt"

	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/telemetry/types"
	"github.com/uptrace/bun"
)

type TelemetryRepository interface {
	CreateInstallEvent(event *types.CliInstallation) error
}

type TelemetryStorage struct {
	DB     *bun.DB
	Ctx    context.Context
	Logger *logger.Logger // optional; nil disables storage logs
}

func NewTelemetryStorage(db *bun.DB, ctx context.Context, l *logger.Logger) *TelemetryStorage {
	return &TelemetryStorage{
		DB:     db,
		Ctx:    ctx,
		Logger: l,
	}
}

func (s *TelemetryStorage) CreateInstallEvent(event *types.CliInstallation) error {
	data := fmt.Sprintf("event_id=%s event_type=%s os=%s arch=%s version=%s", event.ID, event.EventType, event.OS, event.Arch, event.Version)
	storageLog(s.Logger, logger.Debug, "telemetry storage: CreateInstallEvent", data)
	_, err := s.DB.NewInsert().Model(event).Exec(s.Ctx)
	if err != nil {
		storageLog(s.Logger, logger.Error, fmt.Sprintf("telemetry storage: CreateInstallEvent: %v", err), data)
	}
	return err
}
