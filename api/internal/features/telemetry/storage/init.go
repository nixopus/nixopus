package storage

import (
	"context"

	"github.com/nixopus/nixopus/api/internal/features/telemetry/types"
	"github.com/uptrace/bun"
)

type TelemetryRepository interface {
	CreateInstallEvent(event *types.CliInstallation) error
}

type TelemetryStorage struct {
	DB  *bun.DB
	Ctx context.Context
}

func NewTelemetryStorage(db *bun.DB, ctx context.Context) *TelemetryStorage {
	return &TelemetryStorage{
		DB:  db,
		Ctx: ctx,
	}
}

func (s *TelemetryStorage) CreateInstallEvent(event *types.CliInstallation) error {
	_, err := s.DB.NewInsert().Model(event).Exec(s.Ctx)
	return err
}
