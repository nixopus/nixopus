package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	"github.com/nixopus/nixopus/api/internal/queue"
)

type ProvisionInfoProvider interface {
	GetProvisionInfo(ctx context.Context, orgID uuid.UUID, sshKeyID *uuid.UUID) (*storage.ProvisionInfo, error)
}

type LifecycleExecutor func(ctx context.Context, payload queue.MachineLifecyclePayload) (*queue.MachineLifecycleResult, error)

type LifecycleService struct {
	provisionInfo ProvisionInfoProvider
	executeRPC    LifecycleExecutor
}

func NewLifecycleService(p ProvisionInfoProvider, rpc LifecycleExecutor) *LifecycleService {
	return &LifecycleService{provisionInfo: p, executeRPC: rpc}
}

func (s *LifecycleService) resolveInstance(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (string, string, error) {
	info, err := s.provisionInfo.GetProvisionInfo(ctx, orgID, serverID)
	if err != nil {
		return "", "", fmt.Errorf("failed to resolve machine: %w", err)
	}
	if info == nil || info.ContainerName == "" {
		return "", "", types.ErrMachineNotProvisioned
	}
	return info.ContainerName, info.ServerID, nil
}

func (s *LifecycleService) executeAction(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID, action string) (*queue.MachineLifecycleResult, error) {
	instanceName, rpcServerID, err := s.resolveInstance(ctx, orgID, serverID)
	if err != nil {
		return nil, err
	}

	result, err := s.executeRPC(ctx, queue.MachineLifecyclePayload{
		InstanceName: instanceName,
		Action:       action,
		ServerID:     rpcServerID,
	})
	if err != nil {
		if strings.Contains(err.Error(), "timed out") {
			return nil, types.ErrMachineOperationTimeout
		}
		return nil, fmt.Errorf("machine %s failed: %w", action, err)
	}

	if !result.Success {
		return nil, mapResultError(result.Error)
	}

	return result, nil
}

func (s *LifecycleService) GetStatus(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (*types.MachineStateResponse, error) {
	result, err := s.executeAction(ctx, orgID, serverID, "status")
	if err != nil {
		return nil, err
	}

	var state types.MachineState
	if err := json.Unmarshal(result.Data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse machine status: %w", err)
	}

	return &types.MachineStateResponse{
		Status:  "success",
		Message: "Machine status retrieved",
		Data:    &state,
	}, nil
}

func (s *LifecycleService) Restart(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (*types.MachineActionResponse, error) {
	_, err := s.executeAction(ctx, orgID, serverID, "restart")
	if err != nil {
		return nil, err
	}
	return &types.MachineActionResponse{
		Status:  "success",
		Message: "Machine restart initiated",
	}, nil
}

func (s *LifecycleService) Pause(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (*types.MachineActionResponse, error) {
	_, err := s.executeAction(ctx, orgID, serverID, "pause")
	if err != nil {
		return nil, err
	}
	return &types.MachineActionResponse{
		Status:  "success",
		Message: "Machine paused",
	}, nil
}

func (s *LifecycleService) Resume(ctx context.Context, orgID uuid.UUID, serverID *uuid.UUID) (*types.MachineActionResponse, error) {
	_, err := s.executeAction(ctx, orgID, serverID, "resume")
	if err != nil {
		return nil, err
	}
	return &types.MachineActionResponse{
		Status:  "success",
		Message: "Machine resumed",
	}, nil
}

func mapResultError(errMsg string) error {
	switch {
	case strings.Contains(errMsg, "another operation"):
		return types.ErrMachineOperationLocked
	case strings.Contains(errMsg, "not running"):
		return types.ErrMachineNotRunning
	case strings.Contains(errMsg, "already paused"):
		return types.ErrMachineAlreadyPaused
	case strings.Contains(errMsg, "not paused"):
		return types.ErrMachineNotPaused
	default:
		return fmt.Errorf("machine operation failed: %s", errMsg)
	}
}
