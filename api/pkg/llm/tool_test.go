package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
)

func TestToolRegistry_Register(t *testing.T) {
	r := NewToolRegistry()

	err := r.Register(ToolDefinition{
		Name:        "get_weather",
		Description: "Get weather for a city",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`{"temp":22}`), nil
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if r.Count() != 1 {
		t.Errorf("expected 1 tool, got %d", r.Count())
	}
}

func TestToolRegistry_RegisterEmptyName(t *testing.T) {
	r := NewToolRegistry()
	err := r.Register(ToolDefinition{
		Name:    "",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) { return nil, nil },
	})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestToolRegistry_RegisterNilHandler(t *testing.T) {
	r := NewToolRegistry()
	err := r.Register(ToolDefinition{
		Name:    "test",
		Handler: nil,
	})
	if err == nil {
		t.Fatal("expected error for nil handler")
	}
}

func TestToolRegistry_RegisterDuplicate(t *testing.T) {
	r := NewToolRegistry()
	handler := func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) { return nil, nil }

	r.Register(ToolDefinition{Name: "test", Handler: handler})
	err := r.Register(ToolDefinition{Name: "test", Handler: handler})
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}
}

func TestToolRegistry_RegisterNilParameters(t *testing.T) {
	r := NewToolRegistry()
	err := r.Register(ToolDefinition{
		Name:       "no_params",
		Parameters: nil,
		Handler:    func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) { return nil, nil },
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tool, ok := r.Get("no_params")
	if !ok {
		t.Fatal("tool not found")
	}
	if string(tool.Parameters) != `{"type":"object","properties":{}}` {
		t.Errorf("unexpected default parameters: %s", tool.Parameters)
	}
}

func TestToolRegistry_Get(t *testing.T) {
	r := NewToolRegistry()
	r.Register(ToolDefinition{
		Name:    "exists",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) { return nil, nil },
	})

	_, ok := r.Get("exists")
	if !ok {
		t.Error("expected tool to be found")
	}

	_, ok = r.Get("nope")
	if ok {
		t.Error("expected tool to not be found")
	}
}

func TestToolRegistry_Execute(t *testing.T) {
	r := NewToolRegistry()
	r.Register(ToolDefinition{
		Name:        "add",
		Description: "Add two numbers",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"a":{"type":"number"},"b":{"type":"number"}}}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				A float64 `json:"a"`
				B float64 `json:"b"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, err
			}
			result := map[string]float64{"result": input.A + input.B}
			return json.Marshal(result)
		},
	})

	result, err := r.Execute(context.Background(), "add", json.RawMessage(`{"a":3,"b":4}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output map[string]float64
	json.Unmarshal(result, &output)
	if output["result"] != 7 {
		t.Errorf("expected 7, got %f", output["result"])
	}
}

func TestToolRegistry_ExecuteNotFound(t *testing.T) {
	r := NewToolRegistry()
	_, err := r.Execute(context.Background(), "missing", nil)
	if err == nil {
		t.Fatal("expected error for missing tool")
	}
}

func TestToolRegistry_ExecuteWithContext(t *testing.T) {
	r := NewToolRegistry()
	r.Register(ToolDefinition{
		Name: "ctx_check",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
				return json.RawMessage(`"ok"`), nil
			}
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := r.Execute(ctx, "ctx_check", nil)
	if err == nil {
		t.Fatal("expected context error")
	}
}

func TestToolRegistry_List(t *testing.T) {
	r := NewToolRegistry()
	handler := func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) { return nil, nil }

	r.Register(ToolDefinition{Name: "a", Handler: handler})
	r.Register(ToolDefinition{Name: "b", Handler: handler})
	r.Register(ToolDefinition{Name: "c", Handler: handler})

	list := r.List()
	if len(list) != 3 {
		t.Errorf("expected 3 tools, got %d", len(list))
	}

	names := map[string]bool{}
	for _, tool := range list {
		names[tool.Name] = true
	}
	for _, name := range []string{"a", "b", "c"} {
		if !names[name] {
			t.Errorf("tool %q not in list", name)
		}
	}
}

func TestToolRegistry_ToLLMTools(t *testing.T) {
	r := NewToolRegistry()
	r.Register(ToolDefinition{
		Name:        "weather",
		Description: "Get weather",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`),
		Handler:     func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) { return nil, nil },
	})

	tools := r.ToLLMTools()
	if len(tools) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(tools))
	}
	if tools[0].Type != "function" {
		t.Errorf("expected type 'function', got %s", tools[0].Type)
	}
	if tools[0].Function.Name != "weather" {
		t.Errorf("expected name 'weather', got %s", tools[0].Function.Name)
	}
	if tools[0].Function.Description != "Get weather" {
		t.Errorf("unexpected description: %s", tools[0].Function.Description)
	}
}

func TestToolRegistry_Remove(t *testing.T) {
	r := NewToolRegistry()
	r.Register(ToolDefinition{
		Name:    "removable",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) { return nil, nil },
	})

	if !r.Remove("removable") {
		t.Error("expected Remove to return true")
	}
	if r.Count() != 0 {
		t.Error("expected 0 tools after removal")
	}
	if r.Remove("removable") {
		t.Error("expected Remove to return false for non-existent tool")
	}
}

func TestToolRegistry_ConcurrentAccess(t *testing.T) {
	r := NewToolRegistry()
	handler := func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
		return json.RawMessage(`"ok"`), nil
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("tool_%d", i)
			r.Register(ToolDefinition{Name: name, Handler: handler})
		}(i)
	}
	wg.Wait()

	if r.Count() != 100 {
		t.Errorf("expected 100 tools, got %d", r.Count())
	}

	wg = sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("tool_%d", i)
			r.Execute(context.Background(), name, nil)
		}(i)
	}
	wg.Wait()
}

func TestToolDefinition_ToLLMTool(t *testing.T) {
	td := ToolDefinition{
		Name:        "search",
		Description: "Search the web",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}`),
	}

	tool := td.ToLLMTool()
	if tool.Type != "function" {
		t.Errorf("expected 'function', got %s", tool.Type)
	}
	if tool.Function.Name != "search" {
		t.Errorf("expected 'search', got %s", tool.Function.Name)
	}
	if tool.Function.Description != "Search the web" {
		t.Errorf("unexpected description: %s", tool.Function.Description)
	}
	if string(tool.Function.Parameters) != `{"type":"object","properties":{"query":{"type":"string"}},"required":["query"]}` {
		t.Errorf("unexpected parameters: %s", tool.Function.Parameters)
	}
}
