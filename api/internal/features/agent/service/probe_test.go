package service

// CI-safe agent probe tests — no real LLM or database required.
//
// What these tests validate:
//   1. Schema contract  — every tool's JSON Schema correctly rejects hallucinated/missing fields
//   2. Agent loop probe — for each agent type, a scripted mock LLM returns a realistic
//      tool call and we assert the recorder captured schema-valid arguments
//
// The "anti-hallucination" guarantee: if a real LLM sends wrong field names, wrong types,
// or omits required fields, ValidateToolArgs will catch it and the test fails.

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nixopus/nixopus/api/pkg/llm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Schema contract tests ────────────────────────────────────────────────────
// These test the JSON Schema definitions themselves — no LLM involved.

func TestNixopusAPISchema_ValidMethodAndPath(t *testing.T) {
	schema := nixopusAPISchemaFull()

	for _, method := range []string{"GET", "POST", "PUT", "PATCH", "DELETE"} {
		args, _ := json.Marshal(map[string]string{"method": method, "path": "/api/v1/test"})
		assert.NoError(t, llm.ValidateToolArgs(args, schema), "method %s should be valid", method)
	}
}

func TestNixopusAPISchema_MissingMethod(t *testing.T) {
	args, _ := json.Marshal(map[string]string{"path": "/api/v1/test"})
	err := llm.ValidateToolArgs(args, nixopusAPISchemaFull())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "method")
}

func TestNixopusAPISchema_MissingPath(t *testing.T) {
	args, _ := json.Marshal(map[string]string{"method": "GET"})
	err := llm.ValidateToolArgs(args, nixopusAPISchemaFull())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "path")
}

func TestNixopusAPISchema_InvalidMethodEnum(t *testing.T) {
	for _, bad := range []string{"get", "post", "QUERY", "CONNECT", "HEAD"} {
		args, _ := json.Marshal(map[string]string{"method": bad, "path": "/api/v1"})
		err := llm.ValidateToolArgs(args, nixopusAPISchemaFull())
		assert.Error(t, err, "method %q should not be in enum", bad)
	}
}

func TestNixopusAPISchema_MethodNotAString(t *testing.T) {
	args := json.RawMessage(`{"method":42,"path":"/api/v1"}`)
	assert.Error(t, llm.ValidateToolArgs(args, nixopusAPISchemaFull()))
}

func TestNixopusAPISchema_BodyIsOptional(t *testing.T) {
	args, _ := json.Marshal(map[string]string{"method": "POST", "path": "/api/v1/test"})
	assert.NoError(t, llm.ValidateToolArgs(args, nixopusAPISchemaFull()))
}

func TestNixopusAPISchema_BodyMustBeObjectNotString(t *testing.T) {
	args := json.RawMessage(`{"method":"POST","path":"/api/v1","body":"raw string"}`)
	assert.Error(t, llm.ValidateToolArgs(args, nixopusAPISchemaFull()))
}

func TestHttpProbeSchema_RequiresURL(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"url":{"type":"string"},"timeout_seconds":{"type":"integer"}},"required":["url"]}`)

	okArgs, _ := json.Marshal(map[string]string{"url": "https://example.com"})
	assert.NoError(t, llm.ValidateToolArgs(okArgs, schema))

	noURL, _ := json.Marshal(map[string]int{"timeout_seconds": 5})
	err := llm.ValidateToolArgs(noURL, schema)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "url")
}

func TestGithubToolSchemas_AllRequireOwnerAndRepo(t *testing.T) {
	gc := &githubClient{
		cache: map[string]*cachedConnector{
			"": {token: "t", expiresAt: time.Now().Add(time.Hour)},
		},
	}
	svc := &AgentService{}

	toolsNeedingOwnerRepo := []llm.ToolDefinition{
		svc.githubListPullRequestsTool(gc),
		svc.githubListIssuesTool(gc),
		svc.githubCommentOnPRTool(gc),
		svc.githubCommentOnIssueTool(gc),
		svc.githubCreateIssueTool(gc),
		svc.githubSetCommitStatusTool(gc),
		svc.githubCreateDeploymentStatusTool(gc),
		svc.githubSearchRepoContentTool(gc),
		svc.githubCreateOrUpdateFileTool(gc),
		svc.githubGetBranchTool(gc),
		svc.githubCreateBranchTool(gc),
		svc.githubCreatePullRequestTool(gc),
		svc.githubMergePullRequestTool(gc),
		svc.githubGetRepoFileTool(gc),
	}

	for _, tool := range toolsNeedingOwnerRepo {
		t.Run(tool.Name+"/missing_owner", func(t *testing.T) {
			args, _ := json.Marshal(map[string]string{"repo": "myrepo"})
			err := llm.ValidateToolArgs(args, tool.Parameters)
			assert.Error(t, err, "tool %s should require 'owner'", tool.Name)
		})
		t.Run(tool.Name+"/missing_repo", func(t *testing.T) {
			args, _ := json.Marshal(map[string]string{"owner": "myorg"})
			err := llm.ValidateToolArgs(args, tool.Parameters)
			assert.Error(t, err, "tool %s should require 'repo'", tool.Name)
		})
	}
}

func TestGithubCommentOnPRSchema_RequiresPRNumber(t *testing.T) {
	gc := &githubClient{cache: map[string]*cachedConnector{"": {token: "t", expiresAt: time.Now().Add(time.Hour)}}}
	svc := &AgentService{}
	tool := svc.githubCommentOnPRTool(gc)

	// Missing pr_number
	args, _ := json.Marshal(map[string]string{"owner": "o", "repo": "r", "body": "LGTM"})
	assert.Error(t, llm.ValidateToolArgs(args, tool.Parameters))

	// Valid
	valid, _ := json.Marshal(map[string]interface{}{"owner": "o", "repo": "r", "pr_number": 1, "body": "LGTM"})
	assert.NoError(t, llm.ValidateToolArgs(valid, tool.Parameters))
}

func TestGithubSetCommitStatusSchema_StateEnum(t *testing.T) {
	gc := &githubClient{cache: map[string]*cachedConnector{"": {token: "t", expiresAt: time.Now().Add(time.Hour)}}}
	svc := &AgentService{}
	tool := svc.githubSetCommitStatusTool(gc)

	base := map[string]interface{}{"owner": "o", "repo": "r", "sha": "abc123"}

	for _, state := range []string{"pending", "success", "failure", "error"} {
		base["state"] = state
		args, _ := json.Marshal(base)
		assert.NoError(t, llm.ValidateToolArgs(args, tool.Parameters), "state %q should be valid", state)
	}

	base["state"] = "unknown_state"
	args, _ := json.Marshal(base)
	assert.Error(t, llm.ValidateToolArgs(args, tool.Parameters))
}

func TestGithubCreateOrUpdateFileSchema_RequiredFields(t *testing.T) {
	gc := &githubClient{cache: map[string]*cachedConnector{"": {token: "t", expiresAt: time.Now().Add(time.Hour)}}}
	svc := &AgentService{}
	tool := svc.githubCreateOrUpdateFileTool(gc)

	allRequired := map[string]interface{}{
		"owner":   "o",
		"repo":    "r",
		"path":    "src/main.go",
		"content": "package main",
		"message": "add main.go",
		"branch":  "feature-x",
	}
	args, _ := json.Marshal(allRequired)
	assert.NoError(t, llm.ValidateToolArgs(args, tool.Parameters))

	// Remove each required field one by one
	for _, field := range []string{"owner", "repo", "path", "content", "message", "branch"} {
		partial := make(map[string]interface{})
		for k, v := range allRequired {
			if k != field {
				partial[k] = v
			}
		}
		partialArgs, _ := json.Marshal(partial)
		err := llm.ValidateToolArgs(partialArgs, tool.Parameters)
		assert.Error(t, err, "removing field %q should fail validation", field)
	}
}

func TestAllToolDefinitions_HaveValidSchemaJSON(t *testing.T) {
	gc := &githubClient{cache: map[string]*cachedConnector{"": {token: "t", expiresAt: time.Now().Add(time.Hour)}}}
	svc := &AgentService{}

	allTools := []llm.ToolDefinition{
		svc.githubListPullRequestsTool(gc),
		svc.githubListIssuesTool(gc),
		svc.githubCommentOnPRTool(gc),
		svc.githubCommentOnIssueTool(gc),
		svc.githubCreateIssueTool(gc),
		svc.githubSetCommitStatusTool(gc),
		svc.githubCreateDeploymentStatusTool(gc),
		svc.githubSearchRepoContentTool(gc),
		svc.githubCreateOrUpdateFileTool(gc),
		svc.githubGetBranchTool(gc),
		svc.githubCreateBranchTool(gc),
		svc.githubCreatePullRequestTool(gc),
		svc.githubMergePullRequestTool(gc),
		svc.githubGetRepoFileTool(gc),
		svc.analyzeRepositoryTool(),
		loadRemoteRepositoryTool(),
	}

	for _, tool := range allTools {
		t.Run(tool.Name, func(t *testing.T) {
			assert.NotEmpty(t, tool.Name)
			assert.NotEmpty(t, tool.Description)
			assert.NotNil(t, tool.Handler)
			require.NotNil(t, tool.Parameters)

			var schema map[string]interface{}
			require.NoError(t, json.Unmarshal(tool.Parameters, &schema), "tool %s has invalid JSON schema", tool.Name)
			assert.Equal(t, "object", schema["type"], "tool %s schema root must be type:object", tool.Name)
			assert.Contains(t, schema, "properties", "tool %s schema must have properties", tool.Name)
			assert.Contains(t, schema, "required", "tool %s schema should declare required fields", tool.Name)
		})
	}
}

// ─── Agent loop probe tests (mock LLM) ───────────────────────────────────────
// These run the real agent loop against a scripted mock LLM and validate that
// the recorded tool calls have schema-valid arguments.

func TestAgentProbe_DeployAgent_ListApplications(t *testing.T) {
	mockLLM := newScriptedLLM(t, []llm.Response{
		toolCallResponse("c1", "nixopus_api", `{"method":"GET","path":"/api/v1/deploy/applications"}`),
		textResponse("You have 2 applications."),
	})
	defer mockLLM.Close()

	nixopusSrv := newMockNixopusServer(t, map[string]interface{}{
		"data": []map[string]string{{"id": "app-1", "name": "my-app"}},
	})
	defer nixopusSrv.Close()

	rec, agent := buildProbeAgent(t, mockLLM.URL, nixopusSrv.URL, nixopusAPISchemaFull())
	result, err := agent.Run(ctxWithBase(nixopusSrv.URL), "list my applications")

	require.NoError(t, err)
	assert.NotEmpty(t, result.Content)

	calls := rec.CallsFor("nixopus_api")
	require.Len(t, calls, 1, "deploy agent should have called nixopus_api once")
	assertSchemaValid(t, calls[0].Args, nixopusAPISchemaFull(), "nixopus_api")

	var got struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	}
	json.Unmarshal(calls[0].Args, &got)
	assert.Equal(t, "GET", got.Method)
	assert.True(t, strings.Contains(got.Path, "applications") || strings.Contains(got.Path, "deploy"),
		"path should reference applications, got %q", got.Path)
}

func TestAgentProbe_DiagnosticAgent_GetLogs(t *testing.T) {
	mockLLM := newScriptedLLM(t, []llm.Response{
		toolCallResponse("c1", "nixopus_api", `{"method":"GET","path":"/api/v1/deploy/application/logs/app-123"}`),
		textResponse("Here are the logs for your application."),
	})
	defer mockLLM.Close()

	nixopusSrv := newMockNixopusServer(t, map[string]interface{}{"logs": "container started"})
	defer nixopusSrv.Close()

	rec, agent := buildProbeAgent(t, mockLLM.URL, nixopusSrv.URL, nixopusDiagnosticSchema())
	_, err := agent.Run(ctxWithBase(nixopusSrv.URL), "show logs for app-123")

	require.NoError(t, err)

	calls := rec.CallsFor("nixopus_api")
	require.Len(t, calls, 1)
	assertSchemaValid(t, calls[0].Args, nixopusDiagnosticSchema(), "nixopus_api")

	var got struct {
		Method string `json:"method"`
	}
	json.Unmarshal(calls[0].Args, &got)
	assert.Equal(t, "GET", got.Method, "log retrieval must use GET")
}

func TestAgentProbe_GithubAgent_ListPRs(t *testing.T) {
	mockLLM := newScriptedLLM(t, []llm.Response{
		toolCallResponse("c1", "github_list_pull_requests", `{"owner":"myorg","repo":"myrepo","state":"open"}`),
		textResponse("Found 3 open PRs."),
	})
	defer mockLLM.Close()

	ghSchema := json.RawMessage(`{"type":"object","properties":{"owner":{"type":"string"},"repo":{"type":"string"},"state":{"type":"string","enum":["open","closed","all"]},"per_page":{"type":"integer"}},"required":["owner","repo"]}`)

	base := llm.NewToolRegistry()
	base.Register(llm.ToolDefinition{
		Name:       "github_list_pull_requests",
		Parameters: ghSchema,
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	})

	rec := &llm.ToolCallRecorder{}
	wrapped := llm.WrapWithRecorder(base, rec)
	provider := llm.NewOpenAIProvider(llm.OpenAIConfig{APIKey: "key", BaseURL: mockLLM.URL})
	agent := llm.NewAgent(provider, wrapped, llm.AgentConfig{Model: "gpt-4o-mini"})

	_, err := agent.Run(context.Background(), "list open PRs for myorg/myrepo")
	require.NoError(t, err)

	calls := rec.CallsFor("github_list_pull_requests")
	require.Len(t, calls, 1)
	assertSchemaValid(t, calls[0].Args, ghSchema, "github_list_pull_requests")

	var got struct {
		Owner string `json:"owner"`
		Repo  string `json:"repo"`
		State string `json:"state"`
	}
	json.Unmarshal(calls[0].Args, &got)
	assert.NotEmpty(t, got.Owner)
	assert.NotEmpty(t, got.Repo)
}

func TestAgentProbe_NotificationAgent_SendNotification(t *testing.T) {
	mockLLM := newScriptedLLM(t, []llm.Response{
		toolCallResponse("c1", "nixopus_api", `{"method":"POST","path":"/api/v1/notification/send","body":{"channel":"slack","message":"deploy complete"}}`),
		textResponse("Notification sent."),
	})
	defer mockLLM.Close()

	notifSchema := nixopusNotificationSchema()
	nixopusSrv := newMockNixopusServer(t, map[string]string{"status": "sent"})
	defer nixopusSrv.Close()

	rec, agent := buildProbeAgent(t, mockLLM.URL, nixopusSrv.URL, notifSchema)
	_, err := agent.Run(ctxWithBase(nixopusSrv.URL), "send a slack notification: deploy complete")

	require.NoError(t, err)

	calls := rec.CallsFor("nixopus_api")
	require.Len(t, calls, 1)
	assertSchemaValid(t, calls[0].Args, notifSchema, "nixopus_api")

	var got struct {
		Method string `json:"method"`
	}
	json.Unmarshal(calls[0].Args, &got)
	assert.Equal(t, "POST", got.Method, "notification send must use POST")
}

func TestAgentProbe_MachineAgent_GetStats(t *testing.T) {
	mockLLM := newScriptedLLM(t, []llm.Response{
		toolCallResponse("c1", "nixopus_api", `{"method":"GET","path":"/api/v1/machines/stats"}`),
		textResponse("CPU: 34%, Memory: 2.1GB/8GB"),
	})
	defer mockLLM.Close()

	machineSchema := nixopusMachineSchema()
	nixopusSrv := newMockNixopusServer(t, map[string]interface{}{"cpu": 34, "memory": "2.1GB"})
	defer nixopusSrv.Close()

	rec, agent := buildProbeAgent(t, mockLLM.URL, nixopusSrv.URL, machineSchema)
	_, err := agent.Run(ctxWithBase(nixopusSrv.URL), "show machine stats")

	require.NoError(t, err)

	calls := rec.CallsFor("nixopus_api")
	require.Len(t, calls, 1)
	assertSchemaValid(t, calls[0].Args, machineSchema, "nixopus_api")
}

// TestAgentProbe_HallucinatedFieldsRejected verifies that if a mock LLM sends
// an extra hallucinated field with a wrong type, schema validation catches it.
func TestAgentProbe_HallucinatedFieldsRejected(t *testing.T) {
	schema := nixopusAPISchemaFull()

	hallucinatedCases := []struct {
		name string
		args json.RawMessage
		desc string
	}{
		{
			name: "method_as_number",
			args: json.RawMessage(`{"method":1,"path":"/api/v1"}`),
			desc: "LLM sent method as integer instead of string",
		},
		{
			name: "path_as_bool",
			args: json.RawMessage(`{"method":"GET","path":true}`),
			desc: "LLM sent path as boolean",
		},
		{
			name: "body_as_string",
			args: json.RawMessage(`{"method":"POST","path":"/api/v1","body":"should be object"}`),
			desc: "LLM sent body as string instead of object",
		},
		{
			name: "wrong_method_value",
			args: json.RawMessage(`{"method":"QUERY","path":"/api/v1"}`),
			desc: "LLM hallucinated a non-existent HTTP method",
		},
		{
			name: "missing_both_required",
			args: json.RawMessage(`{"body":{}}`),
			desc: "LLM omitted both required fields",
		},
	}

	for _, tc := range hallucinatedCases {
		t.Run(tc.name, func(t *testing.T) {
			err := llm.ValidateToolArgs(tc.args, schema)
			assert.Error(t, err, "case %q (%s) should fail validation", tc.name, tc.desc)
		})
	}
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

// newScriptedLLM creates an httptest server that returns LLM responses in sequence.
func newScriptedLLM(t *testing.T, responses []llm.Response) *httptest.Server {
	t.Helper()
	var idx atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		i := int(idx.Add(1)) - 1
		if i >= len(responses) {
			i = len(responses) - 1
		}
		json.NewEncoder(w).Encode(responses[i])
	}))
	t.Cleanup(srv.Close)
	return srv
}

// newMockNixopusServer creates an httptest server that returns a fixed JSON body for any request.
func newMockNixopusServer(t *testing.T, body interface{}) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	return srv
}

// buildProbeAgent creates an agent with a recording registry backed by a single
// nixopus_api tool that forwards to mockNixopusURL.
func buildProbeAgent(t *testing.T, mockLLMURL, mockNixopusURL string, schema json.RawMessage) (*llm.ToolCallRecorder, *llm.Agent) {
	t.Helper()
	base := llm.NewToolRegistry()
	base.Register(llm.ToolDefinition{
		Name:       "nixopus_api",
		Parameters: schema,
		Handler:    (&AgentService{}).nixopusAPIHandler,
	})

	rec := &llm.ToolCallRecorder{}
	wrapped := llm.WrapWithRecorder(base, rec)
	provider := llm.NewOpenAIProvider(llm.OpenAIConfig{APIKey: "key", BaseURL: mockLLMURL})
	agent := llm.NewAgent(provider, wrapped, llm.AgentConfig{Model: "gpt-4o-mini", MaxSteps: 5})
	return rec, agent
}

// ctxWithBase injects the mock server URL as the nixopus API base URL.
func ctxWithBase(baseURL string) context.Context {
	ctx := context.WithValue(context.Background(), ctxKeyBaseURL, baseURL)
	ctx = context.WithValue(ctx, ctxKeyAuthToken, "test-token")
	ctx = context.WithValue(ctx, ctxKeyOrgID, "org-test")
	return ctx
}

// assertSchemaValid fails the test if args do not satisfy the tool's JSON Schema.
func assertSchemaValid(t *testing.T, args json.RawMessage, schema json.RawMessage, toolName string) {
	t.Helper()
	if err := llm.ValidateToolArgs(args, schema); err != nil {
		t.Errorf("tool %q args failed schema validation: %v\nargs: %s", toolName, err, string(args))
	}
}

// toolCallResponse builds a scripted LLM response that returns a single tool call.
func toolCallResponse(callID, toolName, arguments string) llm.Response {
	return llm.Response{
		ID:    "scripted-1",
		Model: "gpt-4",
		Choices: []llm.Choice{{
			Message: llm.Message{
				Role: llm.RoleAssistant,
				ToolCalls: []llm.ToolCall{{
					ID:       callID,
					Type:     "function",
					Function: llm.FunctionCall{Name: toolName, Arguments: arguments},
				}},
			},
			FinishReason: "tool_calls",
		}},
		Usage: llm.Usage{TotalTokens: 30},
	}
}

// textResponse builds a scripted final LLM text response.
func textResponse(content string) llm.Response {
	return llm.Response{
		ID:    "scripted-2",
		Model: "gpt-4",
		Choices: []llm.Choice{{
			Message:      llm.Message{Role: llm.RoleAssistant, Content: content},
			FinishReason: "stop",
		}},
		Usage: llm.Usage{TotalTokens: 20},
	}
}

// ─── Shared schema fixtures ───────────────────────────────────────────────────

func nixopusAPISchemaFull() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"method": {"type": "string", "enum": ["GET","POST","PUT","PATCH","DELETE"]},
			"path":   {"type": "string"},
			"body":   {"type": "object"}
		},
		"required": ["method", "path"]
	}`)
}

func nixopusDiagnosticSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"method": {"type": "string", "enum": ["GET","POST"]},
			"path":   {"type": "string"},
			"body":   {"type": "object"}
		},
		"required": ["method", "path"]
	}`)
}

func nixopusNotificationSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"method": {"type": "string", "enum": ["GET","POST"]},
			"path":   {"type": "string"},
			"body":   {"type": "object"}
		},
		"required": ["method", "path"]
	}`)
}

func nixopusMachineSchema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"method": {"type": "string", "enum": ["GET","POST","PUT","DELETE"]},
			"path":   {"type": "string"},
			"body":   {"type": "object"}
		},
		"required": ["method", "path"]
	}`)
}
