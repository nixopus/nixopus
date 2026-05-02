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

// TimescaleQuerier abstracts TimescaleStore methods for testability.
type TimescaleQuerier interface {
	GetMetrics(ctx context.Context, machineName string, orgID uuid.UUID, from, to time.Time, limit int) ([]machine_types.MachineMetricRow, error)
	GetEvents(ctx context.Context, machineName string, orgID uuid.UUID, from, to time.Time, limit int) ([]machine_types.MachineEventRow, error)
	GetSummary(ctx context.Context, machineName string, orgID uuid.UUID, from, to time.Time) (*machine_types.MachineSummary, error)
}

// MetricsService queries TimescaleDB for machine observability data, scoped to org.
type MetricsService struct {
	ts             TimescaleQuerier
	db             *bun.DB
	nameResolverFn func(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (string, error) // nil → production
}

func NewMetricsService(ts *storage.TimescaleStore, db *bun.DB) *MetricsService {
	return &MetricsService{ts: ts, db: db}
}

// NewMetricsServiceWith creates a MetricsService with injectable dependencies for tests.
func NewMetricsServiceWith(ts TimescaleQuerier, nameResolverFn func(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (string, error)) *MetricsService {
	return &MetricsService{ts: ts, nameResolverFn: nameResolverFn}
}

func (s *MetricsService) resolveMachineName(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (string, error) {
	if s.nameResolverFn != nil {
		return s.nameResolverFn(ctx, orgID, serverID)
	}
	var row api_types.UserProvisionDetails
	q := s.db.NewSelect().
		Model(&row).
		Column("lxd_container_name").
		Where("lxd_container_name IS NOT NULL").
		OrderExpr("created_at DESC").
		Limit(1)
	if serverID != nil {
		q = q.Where("ssh_key_id = ?", *serverID)
	} else {
		q = q.Where("organization_id = ?", orgID)
	}
	err := q.Scan(ctx)
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

func (s *MetricsService) GetMetrics(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID, from, to time.Time, limit int) (*machine_types.MachineMetricsResponse, error) {
	machineName, err := s.resolveMachineName(ctx, orgID, serverID)
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

func (s *MetricsService) GetEvents(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID, from, to time.Time, limit int) (*machine_types.MachineEventsResponse, error) {
	machineName, err := s.resolveMachineName(ctx, orgID, serverID)
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

func (s *MetricsService) GetSummary(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID, from, to time.Time) (*machine_types.MachineSummaryResponse, error) {
	machineName, err := s.resolveMachineName(ctx, orgID, serverID)
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
