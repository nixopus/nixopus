package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	machine_types "github.com/nixopus/nixopus/api/internal/features/machine/types"
	api_types "github.com/nixopus/nixopus/api/internal/types"
	"github.com/uptrace/bun"
)

// MetricsService queries TimescaleDB for machine observability data, scoped to org.
type MetricsService struct {
	ts *storage.TimescaleStore
	db *bun.DB
}

func NewMetricsService(ts *storage.TimescaleStore, db *bun.DB) *MetricsService {
	return &MetricsService{ts: ts, db: db}
}

// resolveMachineName looks up the lxd_container_name for the org's provisioned machine.
// This name maps to vm_name in the Timescale schema.
func (s *MetricsService) resolveMachineName(ctx context.Context, orgID uuid.UUID) (string, error) {
	var row api_types.UserProvisionDetails
	err := s.db.NewSelect().
		Model(&row).
		Column("lxd_container_name").
		Where("organization_id = ?", orgID).
		Where("lxd_container_name IS NOT NULL").
		OrderExpr("created_at DESC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("no provisioned machine found for org")
		}
		return "", fmt.Errorf("db lookup: %w", err)
	}
	if row.LXDContainerName == nil || *row.LXDContainerName == "" {
		return "", fmt.Errorf("machine name not yet assigned")
	}
	return *row.LXDContainerName, nil
}

func (s *MetricsService) GetMetrics(ctx context.Context, orgID uuid.UUID, from, to time.Time, limit int) (*machine_types.MachineMetricsResponse, error) {
	machineName, err := s.resolveMachineName(ctx, orgID)
	if err != nil {
		return nil, err
	}
	rows, err := s.ts.GetMetrics(ctx, machineName, orgID, from, to, limit)
	if err != nil {
		return nil, err
	}
	return &machine_types.MachineMetricsResponse{
		Status:  "success",
		Message: "Metrics retrieved successfully",
		Data:    rows,
	}, nil
}

func (s *MetricsService) GetEvents(ctx context.Context, orgID uuid.UUID, from, to time.Time, limit int) (*machine_types.MachineEventsResponse, error) {
	machineName, err := s.resolveMachineName(ctx, orgID)
	if err != nil {
		return nil, err
	}
	rows, err := s.ts.GetEvents(ctx, machineName, orgID, from, to, limit)
	if err != nil {
		return nil, err
	}
	return &machine_types.MachineEventsResponse{
		Status:  "success",
		Message: "Events retrieved successfully",
		Data:    rows,
	}, nil
}

func (s *MetricsService) GetSummary(ctx context.Context, orgID uuid.UUID, from, to time.Time) (*machine_types.MachineSummaryResponse, error) {
	machineName, err := s.resolveMachineName(ctx, orgID)
	if err != nil {
		return nil, err
	}
	summary, err := s.ts.GetSummary(ctx, machineName, orgID, from, to)
	if err != nil {
		return nil, err
	}
	return &machine_types.MachineSummaryResponse{
		Status:  "success",
		Message: "Summary retrieved successfully",
		Data:    *summary,
	}, nil
}
