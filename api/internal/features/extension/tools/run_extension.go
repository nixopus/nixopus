package tools

import (
	"context"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	extension_service "github.com/raghavyuva/nixopus-api/internal/features/extension/service"
	extension_storage "github.com/raghavyuva/nixopus-api/internal/features/extension/storage"
	"github.com/raghavyuva/nixopus-api/internal/features/logger"
	shared_storage "github.com/raghavyuva/nixopus-api/internal/storage"
	shared_types "github.com/raghavyuva/nixopus-api/internal/types"
)

// RunExtensionHandler returns the handler function for running an extension
// Auth middleware is applied automatically during registration, so this handler
// only contains the business logic
func RunExtensionHandler(
	store *shared_storage.Store,
	ctx context.Context,
	l logger.Logger,
) func(context.Context, *mcp.CallToolRequest, RunExtensionInput) (*mcp.CallToolResult, RunExtensionOutput, error) {
	return func(
		toolCtx context.Context,
		req *mcp.CallToolRequest,
		input RunExtensionInput,
	) (*mcp.CallToolResult, RunExtensionOutput, error) {
		storage := extension_storage.ExtensionStorage{DB: store.DB, Ctx: ctx}
		service := extension_service.NewExtensionService(store, ctx, l, &storage)

		variables := input.Variables
		if variables == nil {
			variables = make(map[string]interface{})
		}

		execution, err := service.StartRun(input.ExtensionID, variables)
		if err != nil {
			return nil, RunExtensionOutput{}, err
		}

		// Convert shared types to MCP types to avoid circular references
		mcpExecution := convertToMCPExtensionExecution(*execution)

		return nil, RunExtensionOutput{
			Execution: mcpExecution,
		}, nil
	}
}

// convertToMCPExtensionExecution converts shared_types.ExtensionExecution to MCPExtensionExecution
// removing circular references
func convertToMCPExtensionExecution(exec shared_types.ExtensionExecution) MCPExtensionExecution {
	mcpSteps := make([]MCPExecutionStep, len(exec.Steps))
	for i, step := range exec.Steps {
		mcpSteps[i] = MCPExecutionStep{
			ID:          step.ID,
			ExecutionID: step.ExecutionID,
			StepName:    step.StepName,
			Phase:       step.Phase,
			StepOrder:   step.StepOrder,
			StartedAt:   step.StartedAt,
			CompletedAt: step.CompletedAt,
			Status:      step.Status,
			ExitCode:    step.ExitCode,
			Output:      step.Output,
			CreatedAt:   step.CreatedAt,
		}
	}

	return MCPExtensionExecution{
		ID:             exec.ID,
		ExtensionID:    exec.ExtensionID,
		ServerHostname: exec.ServerHostname,
		VariableValues: exec.VariableValues,
		Status:         exec.Status,
		StartedAt:      exec.StartedAt,
		CompletedAt:    exec.CompletedAt,
		ExitCode:       exec.ExitCode,
		ErrorMessage:   exec.ErrorMessage,
		ExecutionLog:   exec.ExecutionLog,
		LogSeq:         exec.LogSeq,
		CreatedAt:      exec.CreatedAt,
		Steps:          mcpSteps,
	}
}
