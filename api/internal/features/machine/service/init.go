package service

import (
	"bytes"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/nixopus/nixopus/api/internal/features/dashboard"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/internal/features/machine/types"
	sshpkg "github.com/nixopus/nixopus/api/internal/features/ssh"
	shared_storage "github.com/nixopus/nixopus/api/internal/storage"
	cryptossh "golang.org/x/crypto/ssh"
)

type MachineService struct {
	store  *shared_storage.Store
	ctx    context.Context
	logger logger.Logger
}

func NewMachineService(store *shared_storage.Store, ctx context.Context, l logger.Logger) *MachineService {
	return &MachineService{
		store:  store,
		ctx:    ctx,
		logger: l,
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
