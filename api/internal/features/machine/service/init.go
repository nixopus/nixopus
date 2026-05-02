package service

import (
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
)

// MachineRegStore is the minimal storage interface required by MachineService.
// It is satisfied by *storage.RegistrationStorage in production and by a mock in tests.
type MachineRegStore interface {
	GetMachineIsActive(serverID uuid.UUID) (bool, error)
	SetMachineActive(serverID uuid.UUID, active bool) error
	GetProvisionResources(serverID uuid.UUID) (vcpu, memMB, diskGB int, err error)
	UpdateProvisionResources(serverID uuid.UUID, vcpu, memMB, diskGB int) error
}

// SSHCommandRunner is the minimal SSH interface required by MachineService.
// It is satisfied by *sshpkg.SSHManager in production and by a mock in tests.
type SSHCommandRunner interface {
	RunCommand(cmd string) (string, error)
}

type MachineService struct {
	store          *shared_storage.Store
	regStore       MachineRegStore
	ctx            context.Context
	logger         logger.Logger
	sshRunnerFn    func(ctx context.Context) (SSHCommandRunner, error)                                        // nil -> production SSH
	collectStatsFn func(runner SSHCommandRunner) (dashboard.SystemStats, error)                               // nil -> dashboard.CollectSystemStats
	sshExecFn      func(ctx context.Context, command string) (stdout, stderr string, exitCode int, err error) // nil -> production SSH session
}

func NewMachineService(store *shared_storage.Store, ctx context.Context, l logger.Logger, regStore *storage.RegistrationStorage) *MachineService {
	return &MachineService{
		store:    store,
		regStore: regStore,
		ctx:      ctx,
		logger:   l,
	}
}

// getSSHRunner returns the SSHCommandRunner to use. It returns the injected runner
// when set (e.g. in tests), otherwise it resolves one from the context via the
// real SSH package.
func (s *MachineService) getSSHRunner(ctx context.Context) (SSHCommandRunner, error) {
	if s.sshRunnerFn != nil {
		return s.sshRunnerFn(ctx)
	}
	return sshpkg.GetSSHManagerFromContext(ctx)
}

func (s *MachineService) collectSystemStats(runner SSHCommandRunner) (dashboard.SystemStats, error) {
	if s.collectStatsFn != nil {
		return s.collectStatsFn(runner)
	}
	return dashboard.CollectSystemStats(s.logger, dashboard.GetSystemStatsOptions{
		CommandExecutor: runner.RunCommand,
	})
}

func (s *MachineService) GetMachineStatus(ctx context.Context, orgID uuid.UUID) (*types.MachineStateResponse, error) {
	serverIDStr, _ := ctx.Value(shared_types.ServerIDKey).(string)
	if serverID, err := uuid.Parse(serverIDStr); err == nil && s.regStore != nil {
		if active, err := s.regStore.GetMachineIsActive(serverID); err == nil && !active {
			return &types.MachineStateResponse{
				Status:  "success",
				Message: "Machine status retrieved",
				Data:    &types.MachineState{Active: false, State: "Paused"},
			}, nil
		}
	}

	runner, err := s.getSSHRunner(ctx)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("failed to get SSH manager: %s", err.Error()), orgID.String())
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	output, err := runner.RunCommand("cat /proc/uptime")
	if err != nil {
		return &types.MachineStateResponse{
			Status:  "success",
			Message: "Machine status retrieved",
			Data:    &types.MachineState{Active: false, State: "Stopped"},
		}, nil
	}

	s.lazyFillSpecs(ctx, runner, orgID)

	return &types.MachineStateResponse{
		Status:  "success",
		Message: "Machine status retrieved",
		Data:    ParseUptimeToState(output),
	}, nil
}

func (s *MachineService) lazyFillSpecs(ctx context.Context, runner SSHCommandRunner, orgID uuid.UUID) {
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

	stats, err := s.collectSystemStats(runner)
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
	runner, err := s.getSSHRunner(ctx)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("failed to get SSH manager: %s", err.Error()), orgID.String())
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	stats, err := s.collectSystemStats(runner)
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

func (s *MachineService) RestartMachine(ctx context.Context, orgID uuid.UUID) (*types.MachineActionResponse, error) {
	runner, err := s.getSSHRunner(ctx)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("failed to get SSH manager: %s", err.Error()), orgID.String())
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}

	_, _ = runner.RunCommand("sudo reboot")

	return &types.MachineActionResponse{
		Status:  "success",
		Message: "Machine restart initiated",
	}, nil
}

func (s *MachineService) PauseMachine(ctx context.Context, orgID uuid.UUID, serverID uuid.UUID) (*types.MachineActionResponse, error) {
	if err := s.regStore.SetMachineActive(serverID, false); err != nil {
		s.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, fmt.Errorf("failed to pause machine: %w", err)
	}
	return &types.MachineActionResponse{
		Status:  "success",
		Message: "Machine paused",
	}, nil
}

func (s *MachineService) ResumeMachine(ctx context.Context, orgID uuid.UUID, serverID uuid.UUID) (*types.MachineActionResponse, error) {
	if err := s.regStore.SetMachineActive(serverID, true); err != nil {
		s.logger.Log(logger.Error, err.Error(), orgID.String())
		return nil, fmt.Errorf("failed to resume machine: %w", err)
	}
	return &types.MachineActionResponse{
		Status:  "success",
		Message: "Machine resumed",
	}, nil
}

func (s *MachineService) ExecCommand(ctx context.Context, orgID uuid.UUID, command string) (*types.HostExecResponse, error) {
	if s.sshExecFn != nil {
		stdout, stderr, exitCode, err := s.sshExecFn(ctx, command)
		if err != nil {
			return nil, fmt.Errorf("command execution failed: %w", err)
		}
		return &types.HostExecResponse{
			Status:  "success",
			Message: "Command executed successfully",
			Data: types.HostExecData{
				Stdout:   stdout,
				Stderr:   stderr,
				ExitCode: exitCode,
			},
		}, nil
	}
	runner, err := s.getSSHRunner(ctx)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("failed to get SSH manager: %s", err.Error()), orgID.String())
		return nil, fmt.Errorf("failed to connect to server: %w", err)
	}
	stdout, err := runner.RunCommand(command)
	if err != nil {
		return nil, fmt.Errorf("command execution failed: %w", err)
	}

	return &types.HostExecResponse{
		Status:  "success",
		Message: "Command executed successfully",
		Data: types.HostExecData{
			Stdout:   stdout,
			Stderr:   "",
			ExitCode: 0,
		},
	}, nil
}
