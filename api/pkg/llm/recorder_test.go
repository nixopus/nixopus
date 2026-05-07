package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// ─── ToolCallRecorder unit tests ─────────────────────────────────────────────

func TestToolCallRecorder_CapturesCall(t *testing.T) {
	base := NewToolRegistry()
	base.Register(ToolDefinition{
		Name:       "echo",
		Parameters: json.RawMessage(`{"type":"object","properties":{"msg":{"type":"string"}}}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return args, nil
		},
	})

	rec := &ToolCallRecorder{}
	wrapped := WrapWithRecorder(base, rec)

	args := json.RawMessage(`{"msg":"hello"}`)
	wrapped.Execute(context.Background(), "echo", args)

	calls := rec.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Tool != "echo" {
		t.Errorf("expected tool 'echo', got %q", calls[0].Tool)
	}
	if string(calls[0].Args) != `{"msg":"hello"}` {
		t.Errorf("unexpected args: %s", calls[0].Args)
	}
	if calls[0].Err != nil {
		t.Errorf("unexpected error: %v", calls[0].Err)
	}
	if string(calls[0].Result) != `{"msg":"hello"}` {
		t.Errorf("unexpected result: %s", calls[0].Result)
	}
}

func TestToolCallRecorder_CapturesError(t *testing.T) {
	base := NewToolRegistry()
	base.Register(ToolDefinition{
		Name: "fail",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return nil, &toolTestError{"boom"}
		},
	})

	rec := &ToolCallRecorder{}
	wrapped := WrapWithRecorder(base, rec)
	wrapped.Execute(context.Background(), "fail", json.RawMessage(`{}`))

	calls := rec.Calls()
	if len(calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(calls))
	}
	if calls[0].Err == nil || calls[0].Err.Error() != "boom" {
		t.Errorf("expected error 'boom', got %v", calls[0].Err)
	}
}

func TestToolCallRecorder_CallsFor(t *testing.T) {
	base := NewToolRegistry()
	for _, name := range []string{"a", "b"} {
		n := name
		base.Register(ToolDefinition{
			Name: n,
			Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
				return json.RawMessage(`"ok"`), nil
			},
		})
	}

	rec := &ToolCallRecorder{}
	wrapped := WrapWithRecorder(base, rec)
	wrapped.Execute(context.Background(), "a", json.RawMessage(`{}`))
	wrapped.Execute(context.Background(), "b", json.RawMessage(`{}`))
	wrapped.Execute(context.Background(), "a", json.RawMessage(`{}`))

	if len(rec.CallsFor("a")) != 2 {
		t.Errorf("expected 2 calls for 'a', got %d", len(rec.CallsFor("a")))
	}
	if len(rec.CallsFor("b")) != 1 {
		t.Errorf("expected 1 call for 'b', got %d", len(rec.CallsFor("b")))
	}
	if len(rec.CallsFor("missing")) != 0 {
		t.Errorf("expected 0 calls for unknown tool")
	}
}

func TestToolCallRecorder_Reset(t *testing.T) {
	base := NewToolRegistry()
	base.Register(ToolDefinition{
		Name: "x",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`1`), nil
		},
	})

	rec := &ToolCallRecorder{}
	wrapped := WrapWithRecorder(base, rec)
	wrapped.Execute(context.Background(), "x", json.RawMessage(`{}`))
	rec.Reset()

	if len(rec.Calls()) != 0 {
		t.Errorf("expected 0 calls after reset, got %d", len(rec.Calls()))
	}
}

func TestWrapWithRecorder_DoesNotModifyBaseRegistry(t *testing.T) {
	var baseCalled atomic.Int32
	base := NewToolRegistry()
	base.Register(ToolDefinition{
		Name: "counter",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			baseCalled.Add(1)
			return json.RawMessage(`1`), nil
		},
	})

	rec := &ToolCallRecorder{}
	wrapped := WrapWithRecorder(base, rec)

	// Call via wrapped
	wrapped.Execute(context.Background(), "counter", json.RawMessage(`{}`))
	// Call directly on base (should not affect recorder)
	base.Execute(context.Background(), "counter", json.RawMessage(`{}`))

	if baseCalled.Load() != 2 {
		t.Errorf("expected base handler called 2 times, got %d", baseCalled.Load())
	}
	if len(rec.Calls()) != 1 {
		t.Errorf("expected 1 recorded call (only the wrapped call), got %d", len(rec.Calls()))
	}
}

func TestWrapWithRecorder_MultipleTools(t *testing.T) {
	base := NewToolRegistry()
	for _, name := range []string{"tool1", "tool2", "tool3"} {
		n := name
		base.Register(ToolDefinition{
			Name: n,
			Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
				return json.RawMessage(`"` + n + `"`), nil
			},
		})
	}

	rec := &ToolCallRecorder{}
	wrapped := WrapWithRecorder(base, rec)

	if wrapped.Count() != 3 {
		t.Errorf("expected 3 tools in wrapped registry, got %d", wrapped.Count())
	}

	for _, name := range []string{"tool1", "tool2", "tool3"} {
		wrapped.Execute(context.Background(), name, json.RawMessage(`{}`))
	}

	if len(rec.Calls()) != 3 {
		t.Errorf("expected 3 total calls, got %d", len(rec.Calls()))
	}
}

// ─── Agent + Recorder integration ────────────────────────────────────────────

// TestAgent_WithRecorder_NixopusAPIProbe tests the deploy agent loop with a
// scripted mock LLM that returns a nixopus_api tool call. Validates:
//  1. The recorder captures the call
//  2. The args pass schema validation (no hallucinated/missing fields)
//  3. Method and path match the expected values
func TestAgent_WithRecorder_NixopusAPIProbe(t *testing.T) {
	var step atomic.Int32
	llmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := step.Add(1)
		var resp Response
		if n == 1 {
			resp = Response{
				ID:    "1",
				Model: "gpt-4",
				Choices: []Choice{{
					Message: Message{
						Role: RoleAssistant,
						ToolCalls: []ToolCall{{
							ID:   "c1",
							Type: "function",
							Function: FunctionCall{
								Name:      "nixopus_api",
								Arguments: `{"method":"GET","path":"/api/v1/deploy/applications"}`,
							},
						}},
					},
					FinishReason: "tool_calls",
				}},
				Usage: Usage{TotalTokens: 20},
			}
		} else {
			resp = Response{
				ID:    "2",
				Model: "gpt-4",
				Choices: []Choice{{
					Message:      Message{Role: RoleAssistant, Content: "You have 2 applications."},
					FinishReason: "stop",
				}},
				Usage: Usage{TotalTokens: 30},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer llmServer.Close()

	nixopusAPISchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"method": {"type": "string", "enum": ["GET","POST","PUT","PATCH","DELETE"]},
			"path":   {"type": "string"},
			"body":   {"type": "object"}
		},
		"required": ["method", "path"]
	}`)

	base := NewToolRegistry()
	base.Register(ToolDefinition{
		Name:       "nixopus_api",
		Parameters: nixopusAPISchema,
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`{"data":[]}`), nil
		},
	})

	rec := &ToolCallRecorder{}
	wrapped := WrapWithRecorder(base, rec)
	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: llmServer.URL})
	agent := NewAgent(provider, wrapped, AgentConfig{Model: "gpt-4"})

	result, err := agent.Run(context.Background(), "list my applications")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content == "" {
		t.Error("expected non-empty response")
	}

	calls := rec.CallsFor("nixopus_api")
	if len(calls) != 1 {
		t.Fatalf("expected 1 nixopus_api call, got %d", len(calls))
	}

	// Schema validation — catches hallucinated or missing fields
	if err := ValidateToolArgs(calls[0].Args, nixopusAPISchema); err != nil {
		t.Errorf("tool call args failed schema validation: %v", err)
	}

	var args struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	}
	json.Unmarshal(calls[0].Args, &args)
	if args.Method != "GET" {
		t.Errorf("expected method GET, got %q", args.Method)
	}
	if !strings.HasPrefix(args.Path, "/api/v1/") {
		t.Errorf("expected path under /api/v1/, got %q", args.Path)
	}
}

// TestAgent_WithRecorder_GithubToolProbe tests the github agent tool call routing.
// Validates that github_list_pull_requests is called with required fields.
func TestAgent_WithRecorder_GithubToolProbe(t *testing.T) {
	var step atomic.Int32
	llmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := step.Add(1)
		var resp Response
		if n == 1 {
			resp = Response{
				ID:    "1",
				Model: "gpt-4",
				Choices: []Choice{{
					Message: Message{
						Role: RoleAssistant,
						ToolCalls: []ToolCall{{
							ID:   "c1",
							Type: "function",
							Function: FunctionCall{
								Name:      "github_list_pull_requests",
								Arguments: `{"owner":"myorg","repo":"myrepo","state":"open"}`,
							},
						}},
					},
					FinishReason: "tool_calls",
				}},
				Usage: Usage{TotalTokens: 20},
			}
		} else {
			resp = Response{
				ID:    "2",
				Model: "gpt-4",
				Choices: []Choice{{
					Message:      Message{Role: RoleAssistant, Content: "Found 3 open PRs."},
					FinishReason: "stop",
				}},
				Usage: Usage{TotalTokens: 30},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer llmServer.Close()

	ghSchema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"owner":    {"type": "string"},
			"repo":     {"type": "string"},
			"state":    {"type": "string", "enum": ["open","closed","all"]},
			"per_page": {"type": "integer"}
		},
		"required": ["owner", "repo"]
	}`)

	base := NewToolRegistry()
	base.Register(ToolDefinition{
		Name:       "github_list_pull_requests",
		Parameters: ghSchema,
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`[]`), nil
		},
	})

	rec := &ToolCallRecorder{}
	wrapped := WrapWithRecorder(base, rec)
	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: llmServer.URL})
	agent := NewAgent(provider, wrapped, AgentConfig{Model: "gpt-4"})

	_, err := agent.Run(context.Background(), "list open PRs for myorg/myrepo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	calls := rec.CallsFor("github_list_pull_requests")
	if len(calls) != 1 {
		t.Fatalf("expected 1 github_list_pull_requests call, got %d", len(calls))
	}

	if err := ValidateToolArgs(calls[0].Args, ghSchema); err != nil {
		t.Errorf("github tool args failed schema validation: %v", err)
	}

	var args struct {
		Owner string `json:"owner"`
		Repo  string `json:"repo"`
	}
	json.Unmarshal(calls[0].Args, &args)
	if args.Owner == "" {
		t.Error("owner must not be empty")
	}
	if args.Repo == "" {
		t.Error("repo must not be empty")
	}
}

// TestAgent_WithRecorder_HallucinatedToolName verifies the agent loop correctly
// returns an error to the model when a hallucinated tool name is returned.
func TestAgent_WithRecorder_HallucinatedToolName(t *testing.T) {
	var step atomic.Int32
	llmServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := step.Add(1)
		var resp Response
		if n == 1 {
			resp = Response{
				ID:    "1",
				Model: "gpt-4",
				Choices: []Choice{{
					Message: Message{
						Role: RoleAssistant,
						ToolCalls: []ToolCall{{
							ID:   "c1",
							Type: "function",
							Function: FunctionCall{
								Name:      "hallucinated_tool_xyz",
								Arguments: `{"foo":"bar"}`,
							},
						}},
					},
					FinishReason: "tool_calls",
				}},
				Usage: Usage{TotalTokens: 20},
			}
		} else {
			resp = Response{
				ID:    "2",
				Model: "gpt-4",
				Choices: []Choice{{
					Message:      Message{Role: RoleAssistant, Content: "I cannot do that."},
					FinishReason: "stop",
				}},
				Usage: Usage{TotalTokens: 20},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer llmServer.Close()

	base := NewToolRegistry()
	base.Register(ToolDefinition{
		Name: "nixopus_api",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`{}`), nil
		},
	})

	rec := &ToolCallRecorder{}
	wrapped := WrapWithRecorder(base, rec)
	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: llmServer.URL})
	agent := NewAgent(provider, wrapped, AgentConfig{Model: "gpt-4"})

	_, err := agent.Run(context.Background(), "do something")
	if err != nil {
		t.Fatalf("unexpected agent error: %v", err)
	}

	// nixopus_api was never actually called — the hallucinated tool got an error
	apiCalls := rec.CallsFor("nixopus_api")
	if len(apiCalls) != 0 {
		t.Errorf("nixopus_api should not have been called, got %d calls", len(apiCalls))
	}
	// Recorder should have zero successful calls
	if len(rec.Calls()) != 0 {
		t.Errorf("no real tool should have been invoked, got %d recorded calls", len(rec.Calls()))
	}
}

// ─── ValidateToolArgs tests ───────────────────────────────────────────────────

func TestValidateToolArgs_PassesValidArgs(t *testing.T) {
	schema := json.RawMessage(`{
		"type": "object",
		"properties": {
			"method":  {"type": "string", "enum": ["GET","POST"]},
			"path":    {"type": "string"},
			"count":   {"type": "integer"},
			"enabled": {"type": "boolean"},
			"tags":    {"type": "array"},
			"meta":    {"type": "object"}
		},
		"required": ["method", "path"]
	}`)

	args := json.RawMessage(`{"method":"GET","path":"/api/v1","count":5,"enabled":true,"tags":[],"meta":{}}`)
	if err := ValidateToolArgs(args, schema); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateToolArgs_MissingRequiredField(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"method":{"type":"string"},"path":{"type":"string"}},"required":["method","path"]}`)
	args := json.RawMessage(`{"method":"GET"}`)

	err := ValidateToolArgs(args, schema)
	if err == nil {
		t.Fatal("expected error for missing required field 'path'")
	}
	if !strings.Contains(err.Error(), "path") {
		t.Errorf("error should mention 'path': %v", err)
	}
}

func TestValidateToolArgs_AllRequiredMissing(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"owner":{"type":"string"},"repo":{"type":"string"}},"required":["owner","repo"]}`)
	args := json.RawMessage(`{}`)

	err := ValidateToolArgs(args, schema)
	if err == nil {
		t.Fatal("expected error for missing required fields")
	}
}

func TestValidateToolArgs_WrongTypeString(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"name":{"type":"string"}},"required":["name"]}`)
	args := json.RawMessage(`{"name":42}`)

	if err := ValidateToolArgs(args, schema); err == nil {
		t.Error("expected type error for number where string expected")
	}
}

func TestValidateToolArgs_WrongTypeNumber(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"count":{"type":"integer"}},"required":["count"]}`)
	args := json.RawMessage(`{"count":"five"}`)

	if err := ValidateToolArgs(args, schema); err == nil {
		t.Error("expected type error for string where number expected")
	}
}

func TestValidateToolArgs_WrongTypeBoolean(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"enabled":{"type":"boolean"}},"required":["enabled"]}`)
	args := json.RawMessage(`{"enabled":"yes"}`)

	if err := ValidateToolArgs(args, schema); err == nil {
		t.Error("expected type error for string where boolean expected")
	}
}

func TestValidateToolArgs_WrongTypeObject(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"body":{"type":"object"}},"required":["body"]}`)
	args := json.RawMessage(`{"body":"not an object"}`)

	if err := ValidateToolArgs(args, schema); err == nil {
		t.Error("expected type error for string where object expected")
	}
}

func TestValidateToolArgs_WrongTypeArray(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"labels":{"type":"array"}},"required":["labels"]}`)
	args := json.RawMessage(`{"labels":"bug"}`)

	if err := ValidateToolArgs(args, schema); err == nil {
		t.Error("expected type error for string where array expected")
	}
}

func TestValidateToolArgs_EnumViolation(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"method":{"type":"string","enum":["GET","POST","DELETE"]}},"required":["method"]}`)
	args := json.RawMessage(`{"method":"PATCH"}`)

	err := ValidateToolArgs(args, schema)
	if err == nil {
		t.Fatal("expected enum violation error")
	}
	if !strings.Contains(err.Error(), "PATCH") {
		t.Errorf("error should mention the invalid value: %v", err)
	}
}

func TestValidateToolArgs_EnumPasses(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"state":{"type":"string","enum":["open","closed","all"]}},"required":["state"]}`)

	for _, v := range []string{"open", "closed", "all"} {
		args, _ := json.Marshal(map[string]string{"state": v})
		if err := ValidateToolArgs(args, schema); err != nil {
			t.Errorf("expected enum value %q to pass: %v", v, err)
		}
	}
}

func TestValidateToolArgs_NotAnObject(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{}}`)

	for _, bad := range []string{`"string"`, `42`, `true`, `null`, `[1,2]`} {
		if err := ValidateToolArgs(json.RawMessage(bad), schema); err == nil {
			t.Errorf("expected error for non-object args: %s", bad)
		}
	}
}

func TestValidateToolArgs_OptionalFieldsNotValidatedWhenAbsent(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"required_field":{"type":"string"},"optional_field":{"type":"integer"}},"required":["required_field"]}`)
	args := json.RawMessage(`{"required_field":"hello"}`)

	if err := ValidateToolArgs(args, schema); err != nil {
		t.Errorf("optional absent fields should not cause errors: %v", err)
	}
}

func TestValidateToolArgs_EmptyRequiredList(t *testing.T) {
	schema := json.RawMessage(`{"type":"object","properties":{"x":{"type":"string"}}}`)
	if err := ValidateToolArgs(json.RawMessage(`{}`), schema); err != nil {
		t.Errorf("no required fields means empty args is valid: %v", err)
	}
}

func TestValidateToolArgs_InvalidSchemaJSON(t *testing.T) {
	if err := ValidateToolArgs(json.RawMessage(`{}`), json.RawMessage(`{bad json`)); err == nil {
		t.Error("expected error for invalid schema JSON")
	}
}

// toolTestError is a simple error type for recorder tests.
type toolTestError struct{ msg string }

func (e *toolTestError) Error() string { return e.msg }
