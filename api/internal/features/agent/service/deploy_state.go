package service

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
)

// Deploy phases tracked across conversation history.
const (
	phaseContextResolved = "context_resolved"
	phaseRepoAnalyzed    = "repo_analyzed"
	phaseProjectCreated  = "project_created"
	phaseDeployStarted   = "deploy_started"
	phaseStatusChecked   = "status_checked"
	phaseLogsChecked     = "logs_checked"
	phaseAppUpdated      = "app_updated"
)

// toolPhaseMap maps tool names to the deploy phase they represent.
var toolPhaseMap = map[string]string{
	"resolve_context":             phaseContextResolved,
	"analyze_repository":          phaseRepoAnalyzed,
	"load_local_workspace":        phaseRepoAnalyzed,
	"load_remote_repository":      phaseRepoAnalyzed,
	"create_project":              phaseProjectCreated,
	"quick_deploy":                phaseDeployStarted,
	"deploy_project":              phaseDeployStarted,
	"get_application_deployments": phaseStatusChecked,
	"get_deployment_logs":         phaseLogsChecked,
	"update_application":          phaseAppUpdated,
}

type deployState struct {
	ApplicationID string
	DeploymentID  string
	Status        string
	Completed     []string
	CurrentPhase  string
}

// extractDeployState scans conversation history to build the current deploy state.
func extractDeployState(history []memory.StoredMessage) string {
	state := &deployState{}
	seenPhases := map[string]bool{}

	for _, msg := range history {
		if msg.Role == llm.RoleAssistant && len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				if phase, ok := toolPhaseMap[tc.Function.Name]; ok {
					if !seenPhases[phase] {
						seenPhases[phase] = true
						state.Completed = append(state.Completed, phase)
					}
					state.CurrentPhase = phase

					extractIDsFromArgs(tc.Function.Arguments, state)
				}
			}
		}

		if msg.Role == llm.RoleTool && msg.Content != "" {
			extractIDsFromResult(msg.Content, state)
		}
	}

	if len(state.Completed) == 0 {
		return "[deploy-state] no_active_deploy [/deploy-state]"
	}

	var sb strings.Builder
	sb.WriteString("[deploy-state]")
	if state.ApplicationID != "" {
		sb.WriteString(fmt.Sprintf(" applicationId=%s", state.ApplicationID))
	}
	if state.DeploymentID != "" {
		sb.WriteString(fmt.Sprintf(" deploymentId=%s", state.DeploymentID))
	}
	if state.Status != "" {
		sb.WriteString(fmt.Sprintf(" status=%s", state.Status))
	}
	sb.WriteString(fmt.Sprintf(" completed=%s", strings.Join(state.Completed, ",")))
	if state.CurrentPhase != "" {
		sb.WriteString(fmt.Sprintf(" current_phase=%s", state.CurrentPhase))
	}
	sb.WriteString(" [/deploy-state]")
	return sb.String()
}

// extractDeployStateFromMessages works with llm.Message slice (for stream handler).
func extractDeployStateFromMessages(messages []llm.Message) string {
	stored := make([]memory.StoredMessage, 0, len(messages))
	for _, m := range messages {
		stored = append(stored, memory.StoredMessage{
			Role:      m.Role,
			Content:   m.Content,
			ToolCalls: m.ToolCalls,
		})
	}
	return extractDeployState(stored)
}

func extractIDsFromArgs(args string, state *deployState) {
	var parsed map[string]interface{}
	if json.Unmarshal([]byte(args), &parsed) != nil {
		return
	}
	if id, ok := parsed["application_id"].(string); ok && id != "" {
		state.ApplicationID = id
	}
	if id, ok := parsed["id"].(string); ok && id != "" && state.ApplicationID == "" {
		state.ApplicationID = id
	}
}

func extractIDsFromResult(content string, state *deployState) {
	var parsed map[string]interface{}
	if json.Unmarshal([]byte(content), &parsed) != nil {
		return
	}

	// Look for IDs in top-level or nested "data" object
	sources := []map[string]interface{}{parsed}
	if data, ok := parsed["data"].(map[string]interface{}); ok {
		sources = append(sources, data)
	}

	for _, src := range sources {
		if id, ok := src["application_id"].(string); ok && id != "" {
			state.ApplicationID = id
		}
		if id, ok := src["id"].(string); ok && id != "" {
			if state.ApplicationID == "" {
				state.ApplicationID = id
			}
		}
		if id, ok := src["deployment_id"].(string); ok && id != "" {
			state.DeploymentID = id
		}
		if s, ok := src["status"].(string); ok && s != "" {
			state.Status = s
		}
	}
}
