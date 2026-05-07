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

func TestAgentRegistry_Register(t *testing.T) {
	r := NewAgentRegistry()
	agent := NewAgent(nil, NewToolRegistry(), AgentConfig{Model: "gpt-4"})
	r.Register("deploy", agent)

	if r.Count() != 1 {
		t.Errorf("expected 1, got %d", r.Count())
	}
}

func TestAgentRegistry_Get(t *testing.T) {
	r := NewAgentRegistry()
	agent := NewAgent(nil, NewToolRegistry(), AgentConfig{Model: "gpt-4"})
	r.Register("deploy", agent)

	got, ok := r.Get("deploy")
	if !ok {
		t.Fatal("expected agent to be found")
	}
	if got != agent {
		t.Error("expected same agent instance")
	}

	_, ok = r.Get("missing")
	if ok {
		t.Error("expected agent to not be found")
	}
}

func TestAgentRegistry_List(t *testing.T) {
	r := NewAgentRegistry()
	r.Register("a", NewAgent(nil, NewToolRegistry(), AgentConfig{Model: "gpt-4"}))
	r.Register("b", NewAgent(nil, NewToolRegistry(), AgentConfig{Model: "gpt-4"}))

	list := r.List()
	if len(list) != 2 {
		t.Errorf("expected 2, got %d", len(list))
	}
}

func TestDelegateTool_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			ID:      "1",
			Model:   "gpt-4",
			Choices: []Choice{{Message: Message{Role: RoleAssistant, Content: "deployed!"}, FinishReason: "stop"}},
			Usage:   Usage{TotalTokens: 10},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	deployAgent := NewAgent(provider, NewToolRegistry(), AgentConfig{Model: "gpt-4", SystemPrompt: "Deploy things"})

	registry := NewAgentRegistry()
	registry.Register("deploy", deployAgent)

	tool := registry.DelegateTool(nil)
	result, err := tool.Handler(context.Background(), json.RawMessage(`{"agent":"deploy","task":"deploy my app"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output map[string]interface{}
	json.Unmarshal(result, &output)
	if output["content"] != "deployed!" {
		t.Errorf("unexpected content: %v", output["content"])
	}
	if output["agent"] != "deploy" {
		t.Errorf("unexpected agent: %v", output["agent"])
	}
}

func TestDelegateTool_NotFound(t *testing.T) {
	registry := NewAgentRegistry()
	registry.Register("a", NewAgent(nil, NewToolRegistry(), AgentConfig{Model: "gpt-4"}))

	tool := registry.DelegateTool(nil)
	result, err := tool.Handler(context.Background(), json.RawMessage(`{"agent":"missing","task":"do stuff"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output map[string]interface{}
	json.Unmarshal(result, &output)
	if output["error"] == nil {
		t.Error("expected error field")
	}
	if output["available"] == nil {
		t.Error("expected available field")
	}
}

func TestDelegateTool_WithStepLimit(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		resp := Response{
			ID:    "1",
			Model: "gpt-4",
			Choices: []Choice{{
				Message:      Message{Role: RoleAssistant, ToolCalls: []ToolCall{{ID: "c1", Type: "function", Function: FunctionCall{Name: "loop", Arguments: "{}"}}}},
				FinishReason: "tool_calls",
			}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	tools := NewToolRegistry()
	tools.Register(ToolDefinition{
		Name: "loop",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`"ok"`), nil
		},
	})
	subAgent := NewAgent(provider, tools, AgentConfig{Model: "gpt-4", MaxSteps: 25})

	registry := NewAgentRegistry()
	registry.Register("sub", subAgent)

	stepLimits := map[string]int{"sub": 2}
	tool := registry.DelegateTool(stepLimits)

	result, _ := tool.Handler(context.Background(), json.RawMessage(`{"agent":"sub","task":"loop"}`))
	var output map[string]interface{}
	json.Unmarshal(result, &output)

	if output["error"] == nil {
		t.Error("expected error from step limit")
	}
	if !strings.Contains(output["error"].(string), "max steps") {
		t.Errorf("unexpected error: %v", output["error"])
	}
}

func TestDelegateTool_InvalidArgs(t *testing.T) {
	registry := NewAgentRegistry()
	tool := registry.DelegateTool(nil)

	_, err := tool.Handler(context.Background(), json.RawMessage(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid args")
	}
}

// --- Parallel Execution ---

func TestRunParallel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var params CompletionParams
		json.NewDecoder(r.Body).Decode(&params)

		content := "response to: " + params.Messages[len(params.Messages)-1].Content
		resp := Response{
			ID:      "1",
			Model:   "gpt-4",
			Choices: []Choice{{Message: Message{Role: RoleAssistant, Content: content}, FinishReason: "stop"}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	agent := NewAgent(provider, NewToolRegistry(), AgentConfig{Model: "gpt-4"})

	tasks := []ParallelTask{
		{Agent: agent, Input: "task1"},
		{Agent: agent, Input: "task2"},
		{Agent: agent, Input: "task3"},
	}

	results := RunParallel(context.Background(), tasks)
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}

	for i, r := range results {
		if r.Err != nil {
			t.Errorf("result %d error: %v", i, r.Err)
		}
		if r.Result == nil {
			t.Errorf("result %d is nil", i)
			continue
		}
		if r.Index != i {
			t.Errorf("expected index %d, got %d", i, r.Index)
		}
	}
}

func TestRunParallel_WithError(t *testing.T) {
	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: "http://127.0.0.1:1"})
	agent := NewAgent(provider, NewToolRegistry(), AgentConfig{Model: "gpt-4"})

	tasks := []ParallelTask{
		{Agent: agent, Input: "will fail"},
	}

	results := RunParallel(context.Background(), tasks)
	if results[0].Err == nil {
		t.Error("expected error")
	}
}

func TestParallelDelegateTool_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			ID:      "1",
			Model:   "gpt-4",
			Choices: []Choice{{Message: Message{Role: RoleAssistant, Content: "done"}, FinishReason: "stop"}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	registry := NewAgentRegistry()
	registry.Register("a", NewAgent(provider, NewToolRegistry(), AgentConfig{Model: "gpt-4"}))
	registry.Register("b", NewAgent(provider, NewToolRegistry(), AgentConfig{Model: "gpt-4"}))

	tool := registry.ParallelDelegateTool()
	result, err := tool.Handler(context.Background(), json.RawMessage(`{"tasks":[{"agent":"a","task":"do X"},{"agent":"b","task":"do Y"}]}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output map[string]interface{}
	json.Unmarshal(result, &output)
	results := output["results"].([]interface{})
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestParallelDelegateTool_AgentNotFound(t *testing.T) {
	registry := NewAgentRegistry()
	registry.Register("a", NewAgent(nil, NewToolRegistry(), AgentConfig{Model: "gpt-4"}))

	tool := registry.ParallelDelegateTool()
	result, _ := tool.Handler(context.Background(), json.RawMessage(`{"tasks":[{"agent":"missing","task":"x"}]}`))

	var output map[string]interface{}
	json.Unmarshal(result, &output)
	if output["error"] == nil {
		t.Error("expected error for missing agent")
	}
}

func TestParallelDelegateTool_InvalidArgs(t *testing.T) {
	registry := NewAgentRegistry()
	tool := registry.ParallelDelegateTool()

	_, err := tool.Handler(context.Background(), json.RawMessage(`{invalid`))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestParallelDelegateTool_WithFailure(t *testing.T) {
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte(`{"error":{"message":"fail"}}`))
	}))
	defer failServer.Close()

	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Response{
			ID: "1", Model: "gpt-4",
			Choices: []Choice{{Message: Message{Role: RoleAssistant, Content: "ok"}, FinishReason: "stop"}},
		})
	}))
	defer okServer.Close()

	registry := NewAgentRegistry()
	registry.Register("fail", NewAgent(NewOpenAIProvider(OpenAIConfig{APIKey: "k", BaseURL: failServer.URL}), NewToolRegistry(), AgentConfig{Model: "gpt-4"}))
	registry.Register("ok", NewAgent(NewOpenAIProvider(OpenAIConfig{APIKey: "k", BaseURL: okServer.URL}), NewToolRegistry(), AgentConfig{Model: "gpt-4"}))

	tool := registry.ParallelDelegateTool()
	result, _ := tool.Handler(context.Background(), json.RawMessage(`{"tasks":[{"agent":"fail","task":"x"},{"agent":"ok","task":"y"}]}`))

	var output map[string]interface{}
	json.Unmarshal(result, &output)
	results := output["results"].([]interface{})

	failResult := results[0].(map[string]interface{})
	if failResult["error"] == nil {
		t.Error("expected error in first result")
	}

	okResult := results[1].(map[string]interface{})
	if okResult["content"] != "ok" {
		t.Errorf("expected 'ok' in second result, got %v", okResult["content"])
	}
}

func TestJoin(t *testing.T) {
	if join(nil, ", ") != "" {
		t.Error("expected empty for nil")
	}
	if join([]string{"a"}, ", ") != "a" {
		t.Error("expected 'a'")
	}
	if join([]string{"a", "b", "c"}, ", ") != "a, b, c" {
		t.Errorf("expected 'a, b, c', got %q", join([]string{"a", "b", "c"}, ", "))
	}
}
