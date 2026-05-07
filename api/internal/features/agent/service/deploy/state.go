package deploy

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/nixopus/nixopus/api/pkg/llm/memory"
)

const (
	PhaseContextResolved = "context_resolved"
	PhaseRepoAnalyzed    = "repo_analyzed"
	PhaseProjectCreated  = "project_created"
	PhaseDeployStarted   = "deploy_started"
	PhaseStatusChecked   = "status_checked"
	PhaseLogsChecked     = "logs_checked"
	PhaseAppUpdated      = "app_updated"
)

// ToolPhaseMap maps tool names to the deploy phase they represent.
var ToolPhaseMap = map[string]string{
	"resolve_context":             PhaseContextResolved,
	"analyze_repository":          PhaseRepoAnalyzed,
	"load_local_workspace":        PhaseRepoAnalyzed,
	"load_remote_repository":      PhaseRepoAnalyzed,
	"create_project":              PhaseProjectCreated,
	"quick_deploy":                PhaseDeployStarted,
	"deploy_project":              PhaseDeployStarted,
	"get_application_deployments": PhaseStatusChecked,
	"get_deployment_logs":         PhaseLogsChecked,
	"update_application":          PhaseAppUpdated,
}

type State struct {
	ApplicationID string
	DeploymentID  string
	Status        string
	Completed     []string
	CurrentPhase  string
}

// ExtractDeployState scans conversation history to build the current deploy state.
func ExtractDeployState(history []memory.StoredMessage) string {
	state := &State{}
	seenPhases := map[string]bool{}

	for _, msg := range history {
		if msg.Role == llm.RoleAssistant && len(msg.ToolCalls) > 0 {
			for _, tc := range msg.ToolCalls {
				if phase, ok := ToolPhaseMap[tc.Function.Name]; ok {
					if !seenPhases[phase] {
						seenPhases[phase] = true
						state.Completed = append(state.Completed, phase)
					}
					state.CurrentPhase = phase

					ExtractIDsFromArgs(tc.Function.Arguments, state)
				}
			}
		}

		if msg.Role == llm.RoleTool && msg.Content != "" {
			ExtractIDsFromResult(msg.Content, state)
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

// ExtractDeployStateFromMessages works with llm.Message slice (for stream handler).
func ExtractDeployStateFromMessages(messages []llm.Message) string {
	stored := make([]memory.StoredMessage, 0, len(messages))
	for _, m := range messages {
		stored = append(stored, memory.StoredMessage{
			Role:      m.Role,
			Content:   m.Content,
			ToolCalls: m.ToolCalls,
		})
	}
	return ExtractDeployState(stored)
}

func ExtractIDsFromArgs(args string, state *State) {
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

func ExtractIDsFromResult(content string, state *State) {
	var parsed map[string]interface{}
	if json.Unmarshal([]byte(content), &parsed) != nil {
		return
	}

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
