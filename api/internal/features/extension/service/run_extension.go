package service

import (
	"encoding/json"

	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	"github.com/raghavyuva/nixopus-api/internal/types"
)

func (s *ExtensionService) StartRun(extensionID string, variableValues map[string]interface{}) (*types.ExtensionExecution, error) {
	ext, err := s.storage.GetExtensionByID(extensionID)
	if err != nil {
		return nil, err
	}

	var spec types.ExtensionSpec
	_ = json.Unmarshal([]byte(ext.ParsedContent), &spec)

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

	go s.executeRun(exec, spec, variableValues)
	return exec, nil
}
