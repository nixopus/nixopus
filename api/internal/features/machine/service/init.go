package service

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/dashboard"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/machine/storage"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	sshpkg "github.com/nixopus/nixopus/api/internal/features/ssh"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	shared_types "github.com/nixopus/nixopus/api/internal/types"
	cryptossh "golang.org/x/crypto/ssh"
)

type MachineService struct {
	store    *shared_storage.Store
	regStore *storage.RegistrationStorage
	ctx      context.Context
	logger   logger.Logger
}

func NewMachineService(store *shared_storage.Store, ctx context.Context, l logger.Logger, regStore *storage.RegistrationStorage) *MachineService {
	return &MachineService{
		store:    store,
		regStore: regStore,
		ctx:      ctx,
		logger:   l,
	}
}

func (s *MachineService) GetMachineStatus(ctx context.Context, orgID uuid.UUID) (*types.MachineStateResponse, error) {
	sshMgr, err := sshpkg.GetSSHManagerFromContext(ctx)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("failed to get SSH manager: %s", err.Error()), orgID.String())
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	output, err := sshMgr.RunCommand("cat /proc/uptime")
	if err != nil {
		return &types.MachineStateResponse{
			Status:  "success",
			Message: "Machine status retrieved",
			Data:    &types.MachineState{Active: false, State: "Stopped"},
		}, nil
	}

	s.lazyFillSpecs(ctx, sshMgr, orgID)

	return &types.MachineStateResponse{
		Status:  "success",
		Message: "Machine status retrieved",
		Data:    ParseUptimeToState(output),
	}, nil
}

func (s *MachineService) lazyFillSpecs(ctx context.Context, sshMgr *sshpkg.SSHManager, orgID uuid.UUID) {
	if s.regStore == nil {
		return
	}

	serverIDStr, _ := ctx.Value(shared_types.ServerIDKey).(string)
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		return
	}

	vcpu, mem, disk, err := s.regStore.GetProvisionResources(serverID)
	if err != nil || (vcpu > 0 && mem > 0 && disk > 0) {
		return
	}

	stats, err := dashboard.CollectSystemStats(s.logger, dashboard.GetSystemStatsOptions{
		CommandExecutor: sshMgr.RunCommand,
	})
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("lazy fill: failed to collect stats: %s", err.Error()), orgID.String())
		return
	}

	cpuCores := stats.CPUCores
	memMB := int(stats.Memory.Total * 1024)
	diskGB := int(stats.Disk.Total)

	if cpuCores == 0 && memMB == 0 && diskGB == 0 {
		return
	}

	if err := s.regStore.UpdateProvisionResources(serverID, cpuCores, memMB, diskGB); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("lazy fill: failed to persist specs: %s", err.Error()), orgID.String())
	}
}

func ParseUptimeToState(raw string) *types.MachineState {
	var uptimeSec float64
	fmt.Sscanf(strings.TrimSpace(raw), "%f", &uptimeSec)
	return &types.MachineState{
		Active:    true,
		State:     "Running",
		UptimeSec: int64(uptimeSec),
	}
}

func (s *MachineService) GetSystemStats(ctx context.Context, orgID uuid.UUID) (*types.SystemStatsResponse, error) {
	sshMgr, err := sshpkg.GetSSHManagerFromContext(ctx)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("failed to get SSH manager: %s", err.Error()), orgID.String())
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	stats, err := dashboard.CollectSystemStats(s.logger, dashboard.GetSystemStatsOptions{
		CommandExecutor: sshMgr.RunCommand,
	})
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("failed to collect system stats: %s", err.Error()), orgID.String())
		return nil, fmt.Errorf("failed to collect system stats: %w", err)
	}

	return &types.SystemStatsResponse{
		Status:  "success",
		Message: "System stats collected successfully",
		Data:    stats,
	}, nil
}

func (s *MachineService) ExecCommand(ctx context.Context, orgID uuid.UUID, command string) (*types.HostExecResponse, error) {
	sshMgr, err := sshpkg.GetSSHManagerFromContext(ctx)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("failed to get SSH manager: %s", err.Error()), orgID.String())
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	session, err := sshMgr.NewSessionWithRetry("")
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	session.Stdout = &stdoutBuf
	session.Stderr = &stderrBuf

	exitCode := 0
	if err := session.Run(command); err != nil {
		if exitErr, ok := err.(*cryptossh.ExitError); ok {
			exitCode = exitErr.ExitStatus()
		} else {
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
	}

	return &types.HostExecResponse{
		Status:  "success",
		Message: "Command executed successfully",
		Data: types.HostExecData{
			Stdout:   stdoutBuf.String(),
			Stderr:   stderrBuf.String(),
			ExitCode: exitCode,
		},
	}, nil
}
