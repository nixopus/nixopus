package service

import (
	"encoding/json"
	"fmt"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/features/ssh"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

// StartRunOptions contains options for starting an extension run
type StartRunOptions struct {
	ExtensionID    string
	VariableValues map[string]interface{}
	Server         *types.Server // Optional: specific server to run on
}

func (s *ExtensionService) StartRun(extensionID string, variableValues map[string]interface{}) (*types.ExtensionExecution, error) {
	return s.StartRunWithServer(StartRunOptions{
		ExtensionID:    extensionID,
		VariableValues: variableValues,
		Server:         nil,
	})
}

func (s *ExtensionService) StartRunWithServer(opts StartRunOptions) (*types.ExtensionExecution, error) {
	ext, err := s.storage.GetExtensionByID(opts.ExtensionID)
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

	varsJSON, _ := json.Marshal(opts.VariableValues)

	// Determine server hostname for the execution record
	serverHostname := ""
	if opts.Server != nil {
		serverHostname = opts.Server.Host
		s.logger.Log(logger.Info, fmt.Sprintf("Running extension on server: %s", serverHostname), "")
	} else {
		serverHostname = "localhost"
		s.logger.Log(logger.Info, "Running extension on default/local server", "")
	}

	exec := &types.ExtensionExecution{
		ExtensionID:    ext.ID,
		ServerHostname: serverHostname,
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

	// Create SSH client based on server configuration
	var sshClient *ssh.SSH
	if opts.Server != nil {
		sshClient = ssh.NewSSHWithServerID(s.store.DB, s.ctx, opts.Server.ID.String())
	} else {
		sshClient = ssh.NewSSH()
	}

	ctx := NewRunContext(exec, spec, opts.VariableValues, sshClient, steps)
	go s.executeRun(ctx)
	return exec, nil
}
