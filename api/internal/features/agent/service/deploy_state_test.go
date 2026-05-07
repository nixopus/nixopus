package service

import (
	"testing"

	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
	"github.com/stretchr/testify/assert"
)

func TestExtractDeployState_NoHistory(t *testing.T) {
	result := extractDeployState(nil)
	assert.Equal(t, "[deploy-state] no_active_deploy [/deploy-state]", result)
}

func TestExtractDeployState_EmptyHistory(t *testing.T) {
	result := extractDeployState([]memory.StoredMessage{})
	assert.Equal(t, "[deploy-state] no_active_deploy [/deploy-state]", result)
}

func TestExtractDeployState_WithAnalyzeRepo(t *testing.T) {
	history := []memory.StoredMessage{
		{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{
				{ID: "tc1", Function: llm.FunctionCall{Name: "analyze_repository", Arguments: `{"repository":"owner/repo"}`}},
			},
		},
		{
			Role:       llm.RoleTool,
			Content:    `{"status":"ok"}`,
			ToolCallID: "tc1",
		},
	}

	result := extractDeployState(history)
	assert.Contains(t, result, "repo_analyzed")
	assert.Contains(t, result, "current_phase=repo_analyzed")
}

func TestExtractDeployState_FullDeploy(t *testing.T) {
	history := []memory.StoredMessage{
		{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{
				{ID: "tc1", Function: llm.FunctionCall{Name: "resolve_context", Arguments: `{}`}},
			},
		},
		{Role: llm.RoleTool, Content: `{}`, ToolCallID: "tc1"},
		{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{
				{ID: "tc2", Function: llm.FunctionCall{Name: "analyze_repository", Arguments: `{}`}},
			},
		},
		{Role: llm.RoleTool, Content: `{}`, ToolCallID: "tc2"},
		{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{
				{ID: "tc3", Function: llm.FunctionCall{Name: "quick_deploy", Arguments: `{"application_id":"app-123"}`}},
			},
		},
		{
			Role:       llm.RoleTool,
			Content:    `{"data":{"application_id":"app-123","deployment_id":"dep-456","status":"building"}}`,
			ToolCallID: "tc3",
		},
		{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{
				{ID: "tc4", Function: llm.FunctionCall{Name: "get_application_deployments", Arguments: `{"application_id":"app-123"}`}},
			},
		},
		{
			Role:       llm.RoleTool,
			Content:    `{"data":{"deployment_id":"dep-456","status":"running"}}`,
			ToolCallID: "tc4",
		},
	}

	result := extractDeployState(history)
	assert.Contains(t, result, "applicationId=app-123")
	assert.Contains(t, result, "deploymentId=dep-456")
	assert.Contains(t, result, "status=running")
	assert.Contains(t, result, "context_resolved")
	assert.Contains(t, result, "repo_analyzed")
	assert.Contains(t, result, "deploy_started")
	assert.Contains(t, result, "status_checked")
	assert.Contains(t, result, "current_phase=status_checked")
}

func TestExtractDeployState_CreateProject(t *testing.T) {
	history := []memory.StoredMessage{
		{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{
				{ID: "tc1", Function: llm.FunctionCall{Name: "create_project", Arguments: `{}`}},
			},
		},
		{
			Role:       llm.RoleTool,
			Content:    `{"data":{"id":"proj-789"}}`,
			ToolCallID: "tc1",
		},
	}

	result := extractDeployState(history)
	assert.Contains(t, result, "project_created")
	assert.Contains(t, result, "applicationId=proj-789")
}

func TestExtractDeployStateFromMessages(t *testing.T) {
	messages := []llm.Message{
		{
			Role: llm.RoleAssistant,
			ToolCalls: []llm.ToolCall{
				{ID: "tc1", Function: llm.FunctionCall{Name: "load_remote_repository", Arguments: `{}`}},
			},
		},
		{Role: llm.RoleTool, Content: `{}`, ToolCallID: "tc1"},
	}

	result := extractDeployStateFromMessages(messages)
	assert.Contains(t, result, "repo_analyzed")
}
