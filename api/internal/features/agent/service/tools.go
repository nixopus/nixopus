package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nixopus/nixopus/api/internal/features/agent/service/codebase"
	agentgithub "github.com/nixopus/nixopus/api/internal/features/agent/service/github"
	"github.com/nixopus/nixopus/api/internal/features/agent/service/scheduler"
	"github.com/nixopus/nixopus/api/internal/features/logger"
	"github.com/nixopus/nixopus/api/pkg/llm"
)

// buildToolProfiles creates the profile builder with shared core tools and
// profile-specific addons. Core tools (nixopus_api, read_skill, http_probe)
// are registered once and shared across all agent profiles.
func (s *AgentService) buildToolProfiles() *llm.ToolProfileBuilder {
	nixopusAPIDef := llm.ToolDefinition{
		Name:        "nixopus_api",
		Description: "Call any Nixopus API endpoint. Use for managing applications, deployments, domains, containers, GitHub connectors, machines, and more. See [api-catalog] in the system prompt for all available operations, required fields, and types.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"method": {"type": "string", "enum": ["GET", "POST", "PUT", "PATCH", "DELETE"], "description": "HTTP method"},
				"path": {"type": "string", "description": "API path (e.g. /api/v1/deploy/applications)"},
				"body": {"type": "object", "description": "Request body (for POST/PUT/PATCH)"}
			},
			"required": ["method", "path"]
		}`),
		Handler: s.nixopusAPIHandler,
	}

	httpProbeDef := llm.ToolDefinition{
		Name:        "http_probe",
		Description: "Check if a URL is accessible and return status code, headers, and response time.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"url":{"type":"string","description":"URL to probe"},"timeout_seconds":{"type":"integer","description":"Timeout in seconds (default 10)"}},"required":["url"]}`),
		Handler:     httpProbeHandler,
	}

	deps := codebase.Deps{
		AuthTokenFromCtx: func(ctx context.Context) string {
			v, _ := ctx.Value(ctxKeyAuthToken).(string)
			return v
		},
		OrgIDFromCtx: func(ctx context.Context) string {
			v, _ := ctx.Value(ctxKeyOrgID).(string)
			return v
		},
		BaseURLFromCtx: func(ctx context.Context) string {
			v, _ := ctx.Value(ctxKeyBaseURL).(string)
			return v
		},
		GetEnvOrDefault: getEnvOrDefault,
	}

	schedDeps := scheduler.ToolDeps{
		Store:     s.scheduleStore,
		Scheduler: func() *scheduler.Scheduler { return s.scheduler },
		Memory:    s.memory,
		Logger:    s.logger,
		UserID: func(ctx context.Context) string {
			if uid, ok := ctx.Value(ctxKeySchedulerUserID).(string); ok && uid != "" {
				return uid
			}
			// Scheduled agent runs set internal auth context, not the chat session key.
			if uid, ok := ctx.Value(ctxKeyInternalUserID).(string); ok && uid != "" {
				return uid
			}
			return ""
		},
		OrgID: func(ctx context.Context) string {
			v, _ := ctx.Value(ctxKeyOrgID).(string)
			return v
		},
	}

	// Core tools shared by all profiles
	pb := llm.NewToolProfileBuilder(func(reg *llm.ToolRegistry) {
		reg.Register(s.skills.Tool())
		reg.Register(nixopusAPIDef)
		reg.Register(httpProbeDef)
	})

	// Deploy: core + delegation + codebase analysis + scheduling
	pb.RegisterProfile(llm.ProfileDeploy, func(reg *llm.ToolRegistry) {
		reg.Register(s.agents.DelegateTool(map[string]int{
			"diagnostic":   8,
			"github":       5,
			"notification": 3,
			"machine":      8,
		}))
		reg.Register(codebase.AnalyzeRepositoryTool(deps))
		reg.Register(codebase.LoadRemoteRepositoryTool())
		reg.Register(scheduler.CreateScheduleTool(schedDeps))
		reg.Register(scheduler.ListSchedulesTool(schedDeps))
		reg.Register(scheduler.DeleteScheduleTool(schedDeps))
		reg.Register(scheduler.PauseScheduleTool(schedDeps))
		reg.Register(scheduler.ResumeScheduleTool(schedDeps))
	})

	// Diagnostic: core only (nixopus_api + http_probe + read_skill)

	// GitHub: core + all github_* tools
	gc := s.github
	pb.RegisterProfile(llm.ProfileGitHub, func(reg *llm.ToolRegistry) {
		reg.Register(agentgithub.ListPullRequestsTool(gc))
		reg.Register(agentgithub.ListIssuesTool(gc))
		reg.Register(agentgithub.CommentOnPRTool(gc))
		reg.Register(agentgithub.CommentOnIssueTool(gc))
		reg.Register(agentgithub.CreateIssueTool(gc))
		reg.Register(agentgithub.SetCommitStatusTool(gc))
		reg.Register(agentgithub.CreateDeploymentStatusTool(gc))
		reg.Register(agentgithub.SearchRepoContentTool(gc))
		reg.Register(agentgithub.CreateOrUpdateFileTool(gc))
		reg.Register(agentgithub.GetBranchTool(gc))
		reg.Register(agentgithub.CreateBranchTool(gc))
		reg.Register(agentgithub.CreatePullRequestTool(gc))
		reg.Register(agentgithub.MergePullRequestTool(gc))
		reg.Register(agentgithub.GetRepoFileTool(gc))
	})

	// Notification and Machine profiles use core tools only (no addons)

	return pb
}

// nixopusAPIHandler calls the Nixopus API on behalf of the user.
// Auth context (token, org) is passed via context from the controller.
// It runs a pre-flight validation against the OpenAPI spec BEFORE sending the
// HTTP request so the LLM receives corrective feedback immediately rather than
// having a failed deployment record created in the database.
func (s *AgentService) nixopusAPIHandler(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var input struct {
		Method string          `json:"method"`
		Path   string          `json:"path"`
		Body   json.RawMessage `json:"body,omitempty"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		s.logger.Log(logger.Error, "nixopus_api: failed to unmarshal args", err.Error())
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	s.logger.Log(logger.Debug, "nixopus_api: invoked",
		fmt.Sprintf("method=%s path=%s body_len=%d", input.Method, input.Path, len(input.Body)))

	if s.preflight != nil {
		if errMsg := s.preflight.Validate(input.Method, input.Path, input.Body); errMsg != "" {
			s.logger.Log(logger.Warning, "nixopus_api: preflight validation failed",
				fmt.Sprintf("method=%s path=%s err=%s", input.Method, input.Path, errMsg))
			return json.Marshal(map[string]interface{}{
				"preflight_error": errMsg,
				"action_required": "Fix the listed field errors and retry the nixopus_api call with a corrected body.",
			})
		}
	}

	authToken, _ := ctx.Value(ctxKeyAuthToken).(string)
	orgID, _ := ctx.Value(ctxKeyOrgID).(string)
	baseURL, _ := ctx.Value(ctxKeyBaseURL).(string)

	internalSig, _ := ctx.Value(ctxKeyInternalSig).(string)
	internalUserID, _ := ctx.Value(ctxKeyInternalUserID).(string)

	authMode := "none"
	if internalSig != "" && internalUserID != "" {
		authMode = "internal"
	} else if authToken != "" {
		authMode = "bearer"
	}

	s.logger.Log(logger.Debug, "nixopus_api: auth context",
		fmt.Sprintf("auth_mode=%s org_id=%q base_url=%q has_token=%t internal_user=%q",
			authMode, orgID, baseURL, authToken != "", internalUserID))

	if baseURL == "" {
		baseURL = "http://localhost:" + getEnvOrDefault("PORT", "2089")
	}

	var bodyReader io.Reader
	if input.Body != nil && input.Method != "GET" {
		bodyReader = jsonReader(input.Body)
	}

	req, err := http.NewRequestWithContext(ctx, input.Method, baseURL+input.Path, bodyReader)
	if err != nil {
		s.logger.Log(logger.Error, "nixopus_api: failed to create HTTP request",
			fmt.Sprintf("method=%s url=%s err=%s", input.Method, baseURL+input.Path, err.Error()))
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if internalSig != "" && internalUserID != "" {
		req.Header.Set("X-Nixopus-Internal-Sig", internalSig)
		req.Header.Set("X-Internal-User-Id", internalUserID)
		req.Header.Set("X-Organization-Id", orgID)
	} else if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	if orgID != "" {
		req.Header.Set("X-Organization-ID", orgID)
	}

	start := time.Now()
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		s.logger.Log(logger.Error, "nixopus_api: HTTP call failed",
			fmt.Sprintf("method=%s path=%s auth_mode=%s elapsed=%s err=%s",
				input.Method, input.Path, authMode, elapsed, err.Error()))
		return nil, fmt.Errorf("api call failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 32*1024))

	if resp.StatusCode >= 400 {
		s.logger.Log(logger.Warning, "nixopus_api: HTTP error response",
			fmt.Sprintf("method=%s path=%s status=%d auth_mode=%s elapsed=%s body=%s",
				input.Method, input.Path, resp.StatusCode, authMode, elapsed, truncate(string(body), 500)))
		return json.Marshal(map[string]interface{}{
			"error":       fmt.Sprintf("API returned %d", resp.StatusCode),
			"status_code": resp.StatusCode,
			"body":        string(body),
		})
	}

	s.logger.Log(logger.Debug, "nixopus_api: success",
		fmt.Sprintf("method=%s path=%s status=%d elapsed=%s body_len=%d",
			input.Method, input.Path, resp.StatusCode, elapsed, len(body)))

	var result interface{}
	if json.Unmarshal(body, &result) == nil {
		return json.Marshal(result)
	}
	return json.Marshal(map[string]string{"response": string(body)})
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func httpProbeHandler(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
	var input struct {
		URL            string `json:"url"`
		TimeoutSeconds int    `json:"timeout_seconds"`
	}
	if err := json.Unmarshal(args, &input); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	timeout := time.Duration(input.TimeoutSeconds) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}

	client := &http.Client{Timeout: timeout}
	start := time.Now()

	resp, err := client.Get(input.URL)
	elapsed := time.Since(start)

	if err != nil {
		return json.Marshal(map[string]interface{}{
			"reachable":   false,
			"error":       err.Error(),
			"response_ms": elapsed.Milliseconds(),
		})
	}
	defer resp.Body.Close()

	return json.Marshal(map[string]interface{}{
		"reachable":    true,
		"status_code":  resp.StatusCode,
		"response_ms":  elapsed.Milliseconds(),
		"content_type": resp.Header.Get("Content-Type"),
	})
}

type contextKey string

const (
	ctxKeyAuthToken       contextKey = "auth_token"
	ctxKeyOrgID           contextKey = "org_id"
	ctxKeyBaseURL         contextKey = "base_url"
	ctxKeySchedulerUserID contextKey = "scheduler_user_id"
)

func jsonReader(data json.RawMessage) io.Reader {
	return bytes.NewReader(data)
}
