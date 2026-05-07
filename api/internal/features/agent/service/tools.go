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
	"github.com/nixopus/nixopus/api/pkg/llm"
)

// buildDeployTools creates the core tool registry for the deploy agent.
// Tools call back into the Nixopus API using the user's auth context.
func (s *AgentService) buildDeployTools() *llm.ToolRegistry {
	tools := llm.NewToolRegistry()

	tools.Register(s.skills.Tool())
	tools.Register(s.agents.DelegateTool(map[string]int{
		"diagnostic":   8,
		"github":       5,
		"notification": 3,
		"machine":      8,
	}))

	tools.Register(llm.ToolDefinition{
		Name:        "nixopus_api",
		Description: "Call any Nixopus API endpoint. Use for managing applications, deployments, domains, containers, GitHub connectors, machines, and more.",
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
	})

	tools.Register(llm.ToolDefinition{
		Name:        "http_probe",
		Description: "Check if a URL is accessible and return status code, headers, and response time.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"url":{"type":"string","description":"URL to probe"},"timeout_seconds":{"type":"integer","description":"Timeout in seconds (default 10)"}},"required":["url"]}`),
		Handler:     httpProbeHandler,
	})

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

	tools.Register(codebase.AnalyzeRepositoryTool(deps))
	tools.Register(codebase.LoadRemoteRepositoryTool())

	return tools
}

func (s *AgentService) buildDiagnosticTools() *llm.ToolRegistry {
	tools := llm.NewToolRegistry()
	tools.Register(s.skills.Tool())

	tools.Register(llm.ToolDefinition{
		Name:        "nixopus_api",
		Description: "Call Nixopus API for diagnostic data (logs, healthchecks, container status).",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"method": {"type": "string", "enum": ["GET", "POST"], "description": "HTTP method"},
				"path": {"type": "string", "description": "API path"},
				"body": {"type": "object", "description": "Request body"}
			},
			"required": ["method", "path"]
		}`),
		Handler: s.nixopusAPIHandler,
	})

	tools.Register(llm.ToolDefinition{
		Name:        "http_probe",
		Description: "Probe a URL for health checking.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"url":{"type":"string"},"timeout_seconds":{"type":"integer"}},"required":["url"]}`),
		Handler:     httpProbeHandler,
	})

	return tools
}

func (s *AgentService) buildGithubTools() *llm.ToolRegistry {
	tools := llm.NewToolRegistry()
	tools.Register(s.skills.Tool())

	gc := s.github

	tools.Register(agentgithub.ListPullRequestsTool(gc))
	tools.Register(agentgithub.ListIssuesTool(gc))
	tools.Register(agentgithub.CommentOnPRTool(gc))
	tools.Register(agentgithub.CommentOnIssueTool(gc))
	tools.Register(agentgithub.CreateIssueTool(gc))
	tools.Register(agentgithub.SetCommitStatusTool(gc))
	tools.Register(agentgithub.CreateDeploymentStatusTool(gc))
	tools.Register(agentgithub.SearchRepoContentTool(gc))
	tools.Register(agentgithub.CreateOrUpdateFileTool(gc))
	tools.Register(agentgithub.GetBranchTool(gc))
	tools.Register(agentgithub.CreateBranchTool(gc))
	tools.Register(agentgithub.CreatePullRequestTool(gc))
	tools.Register(agentgithub.MergePullRequestTool(gc))
	tools.Register(agentgithub.GetRepoFileTool(gc))

	tools.Register(llm.ToolDefinition{
		Name:        "nixopus_api",
		Description: "Call Nixopus GitHub connector API.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"method": {"type": "string", "enum": ["GET", "POST", "DELETE"]},
				"path": {"type": "string"},
				"body": {"type": "object"}
			},
			"required": ["method", "path"]
		}`),
		Handler: s.nixopusAPIHandler,
	})

	return tools
}

func (s *AgentService) buildNotificationTools() *llm.ToolRegistry {
	tools := llm.NewToolRegistry()

	tools.Register(llm.ToolDefinition{
		Name:        "nixopus_api",
		Description: "Send notifications via Nixopus API.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"method": {"type": "string", "enum": ["GET", "POST"]},
				"path": {"type": "string"},
				"body": {"type": "object"}
			},
			"required": ["method", "path"]
		}`),
		Handler: s.nixopusAPIHandler,
	})

	return tools
}

func (s *AgentService) buildMachineTools() *llm.ToolRegistry {
	tools := llm.NewToolRegistry()
	tools.Register(s.skills.Tool())

	tools.Register(llm.ToolDefinition{
		Name:        "nixopus_api",
		Description: "Call Nixopus machine/server management API.",
		Parameters: json.RawMessage(`{
			"type": "object",
			"properties": {
				"method": {"type": "string", "enum": ["GET", "POST", "PUT", "DELETE"]},
				"path": {"type": "string"},
				"body": {"type": "object"}
			},
			"required": ["method", "path"]
		}`),
		Handler: s.nixopusAPIHandler,
	})

	return tools
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
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	if s.preflight != nil {
		if errMsg := s.preflight.Validate(input.Method, input.Path, input.Body); errMsg != "" {
			return json.Marshal(map[string]interface{}{
				"preflight_error": errMsg,
				"action_required": "Fix the listed field errors and retry the nixopus_api call with a corrected body.",
			})
		}
	}

	authToken, _ := ctx.Value(ctxKeyAuthToken).(string)
	orgID, _ := ctx.Value(ctxKeyOrgID).(string)
	baseURL, _ := ctx.Value(ctxKeyBaseURL).(string)

	if baseURL == "" {
		baseURL = "http://localhost:" + getEnvOrDefault("PORT", "2089")
	}

	var bodyReader io.Reader
	if input.Body != nil && input.Method != "GET" {
		bodyReader = jsonReader(input.Body)
	}

	req, err := http.NewRequestWithContext(ctx, input.Method, baseURL+input.Path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if authToken != "" {
		req.Header.Set("Authorization", "Bearer "+authToken)
	}
	if orgID != "" {
		req.Header.Set("X-Organization-ID", orgID)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api call failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 32*1024))

	if resp.StatusCode >= 400 {
		return json.Marshal(map[string]interface{}{
			"error":       fmt.Sprintf("API returned %d", resp.StatusCode),
			"status_code": resp.StatusCode,
			"body":        string(body),
		})
	}

	var result interface{}
	if json.Unmarshal(body, &result) == nil {
		return json.Marshal(result)
	}
	return json.Marshal(map[string]string{"response": string(body)})
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
	ctxKeyAuthToken contextKey = "auth_token"
	ctxKeyOrgID     contextKey = "org_id"
	ctxKeyBaseURL   contextKey = "base_url"
)

func jsonReader(data json.RawMessage) io.Reader {
	return bytes.NewReader(data)
}
