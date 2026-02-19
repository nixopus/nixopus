package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *ExtensionService) StartRun(ctx context.Context, extensionID string, variableValues map[string]interface{}) (*types.ExtensionExecution, error) {
	ext, err := s.storage.GetExtensionByID(extensionID)
	if err != nil {
		return nil, err
	}

	s.logger.Log(logger.Info, fmt.Sprintf("Extension ParsedContent: %s", ext.ParsedContent), "")

	var spec types.ExtensionSpec
	var jsonString string
	if err := json.Unmarshal([]byte(ext.ParsedContent), &jsonString); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to unmarshal JSON string: %v", err), "")
		return nil, err
	}

	if err := json.Unmarshal([]byte(jsonString), &spec); err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("Failed to unmarshal extension spec: %v", err), "")
		return nil, err
	}

	s.logger.Log(logger.Info, fmt.Sprintf("Parsed spec - Run steps: %d, Validate steps: %d", len(spec.Execution.Run), len(spec.Execution.Validate)), "")

	varsJSON, _ := json.Marshal(variableValues)
	exec := &types.ExtensionExecution{
		ExtensionID:    ext.ID,
		VariableValues: string(varsJSON),
		Status:         types.ExecutionStatusRunning,
	}
	if err := s.storage.CreateExecution(exec); err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	var steps []types.ExecutionStep
	order := 1
	for _, st := range spec.Execution.Run {
		steps = append(steps, types.ExecutionStep{
			ExecutionID: exec.ID,
			StepName:    st.Name,
			Phase:       "run",
			StepOrder:   order,
			Status:      types.ExecutionStatusPending,
		})
		order++
	}
	for _, st := range spec.Execution.Validate {
		steps = append(steps, types.ExecutionStep{
			ExecutionID: exec.ID,
			StepName:    st.Name,
			Phase:       "validate",
			StepOrder:   order,
			Status:      types.ExecutionStatusPending,
		})
		order++
	}
	if err := s.storage.CreateExecutionSteps(steps); err != nil {
		s.logger.Log(logger.Error, err.Error(), "")
		return nil, err
	}

	sshManager, err := ssh.GetSSHManagerFromContext(ctx)
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("failed to get SSH manager: %v", err), "")
		return nil, err
	}
	sshClient, err := sshManager.GetDefaultSSH()
	if err != nil {
		s.logger.Log(logger.Error, fmt.Sprintf("failed to get default SSH client: %v", err), "")
		return nil, err
	}
	runCtx := NewRunContext(ctx, exec, spec, variableValues, sshClient, steps)
	go s.executeRun(runCtx)
	return exec, nil
}
