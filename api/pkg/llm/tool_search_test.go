package llm

import (
	"context"
	"encoding/json"
	"testing"
)

func TestToolSearch_Add(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{})
	ts.Add(SearchableToolEntry{
		Definition: ToolDefinition{Name: "deploy", Description: "Deploy applications"},
		Keywords:   []string{"deploy", "application", "release"},
	})
	if ts.Count() != 1 {
		t.Errorf("expected 1, got %d", ts.Count())
	}
}

func TestToolSearch_AddTool(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{})
	ts.AddTool(ToolDefinition{Name: "search", Description: "Search the web"}, "search", "web", "query")
	if ts.Count() != 1 {
		t.Errorf("expected 1, got %d", ts.Count())
	}
}

func TestToolSearch_AddAutoKeywords(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{})
	ts.Add(SearchableToolEntry{
		Definition: ToolDefinition{Name: "get_weather", Description: "Get current weather for a city"},
	})
	// Keywords should be auto-extracted from name+description
	results := ts.Search("weather city")
	if len(results) != 1 {
		t.Errorf("expected 1 result from auto-keywords, got %d", len(results))
	}
}

func TestToolSearch_Search(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{TopK: 3, MinScore: 0.1})
	ts.AddTool(ToolDefinition{Name: "deploy_app", Description: "Deploy an application"}, "deploy", "application", "release", "production")
	ts.AddTool(ToolDefinition{Name: "get_logs", Description: "Get application logs"}, "logs", "application", "debug", "errors")
	ts.AddTool(ToolDefinition{Name: "get_weather", Description: "Get weather"}, "weather", "temperature", "forecast")
	ts.AddTool(ToolDefinition{Name: "restart_app", Description: "Restart application"}, "restart", "application", "redeploy")

	results := ts.Search("deploy application to production")
	if len(results) == 0 {
		t.Fatal("expected results for deploy query")
	}
	if results[0].Name != "deploy_app" {
		t.Errorf("expected deploy_app first, got %s", results[0].Name)
	}
}

func TestToolSearch_SearchNoResults(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{MinScore: 0.5})
	ts.AddTool(ToolDefinition{Name: "deploy", Description: "Deploy"}, "deploy")

	results := ts.Search("quantum physics simulation")
	if len(results) != 0 {
		t.Errorf("expected 0 results, got %d", len(results))
	}
}

func TestToolSearch_SearchEmptyQuery(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{})
	ts.AddTool(ToolDefinition{Name: "test", Description: "Test"}, "test")

	results := ts.Search("")
	if results != nil {
		t.Errorf("expected nil for empty query, got %d results", len(results))
	}
}

func TestToolSearch_TopKLimit(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{TopK: 2, MinScore: 0.01})
	for i := 0; i < 10; i++ {
		ts.AddTool(ToolDefinition{Name: "tool", Description: "tool"}, "tool", "common")
	}

	results := ts.Search("tool common")
	if len(results) > 2 {
		t.Errorf("expected max 2 results (TopK), got %d", len(results))
	}
}

func TestToolSearch_DefaultConfig(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{})
	if ts.config.TopK != 6 {
		t.Errorf("expected default TopK 6, got %d", ts.config.TopK)
	}
	if ts.config.MinScore != 0.1 {
		t.Errorf("expected default MinScore 0.1, got %f", ts.config.MinScore)
	}
}

func TestToolSearch_Tool(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{TopK: 2, MinScore: 0.01})
	ts.AddTool(ToolDefinition{
		Name:        "get_weather",
		Description: "Get weather for a city",
		Parameters:  json.RawMessage(`{}`),
		Handler:     func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) { return nil, nil },
	}, "weather", "city", "temperature")

	registry := NewToolRegistry()
	tool := ts.Tool(registry)

	if tool.Name != "load_tool" {
		t.Errorf("unexpected tool name: %s", tool.Name)
	}

	result, err := tool.Handler(context.Background(), json.RawMessage(`{"query":"weather temperature"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var output map[string]interface{}
	json.Unmarshal(result, &output)
	if output["count"] == nil {
		t.Error("expected count in result")
	}

	// Tool should now be registered
	if _, ok := registry.Get("get_weather"); !ok {
		t.Error("expected get_weather to be registered in registry")
	}
}

func TestToolSearch_ToolNoResults(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{MinScore: 0.9})
	ts.AddTool(ToolDefinition{Name: "deploy", Description: "Deploy"}, "deploy")

	registry := NewToolRegistry()
	tool := ts.Tool(registry)

	result, _ := tool.Handler(context.Background(), json.RawMessage(`{"query":"quantum"}`))
	var output map[string]interface{}
	json.Unmarshal(result, &output)
	if output["message"] == nil {
		t.Error("expected 'no matching tools' message")
	}
}

func TestToolSearch_ToolInvalidArgs(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{})
	registry := NewToolRegistry()
	tool := ts.Tool(registry)

	_, err := tool.Handler(context.Background(), json.RawMessage(`{invalid`))
	if err == nil {
		t.Fatal("expected error for invalid args")
	}
}

func TestToolSearch_ToolAlreadyRegistered(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{MinScore: 0.01})
	ts.AddTool(ToolDefinition{
		Name:    "existing",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) { return nil, nil },
	}, "test")

	registry := NewToolRegistry()
	registry.Register(ToolDefinition{
		Name:    "existing",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) { return nil, nil },
	})

	tool := ts.Tool(registry)
	tool.Handler(context.Background(), json.RawMessage(`{"query":"test"}`))

	// Should not error, just skip re-registration
	if registry.Count() != 1 {
		t.Errorf("expected 1 tool (not duplicated), got %d", registry.Count())
	}
}

func TestToolSearch_Processor(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{TopK: 2, MinScore: 0.01})
	ts.AddTool(ToolDefinition{Name: "deploy", Description: "Deploy"}, "deploy", "release")
	ts.AddTool(ToolDefinition{Name: "logs", Description: "Logs"}, "logs", "debug")

	proc := ts.Processor()
	input := &StepInput{
		Messages: []Message{{Role: RoleUser, Content: "deploy my app"}},
		Tools:    []Tool{},
	}

	err := proc.ProcessInput(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(input.Tools) == 0 {
		t.Error("expected tools to be added")
	}
}

func TestToolSearch_ProcessorNoMessages(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{})
	proc := ts.Processor()

	input := &StepInput{Messages: []Message{}}
	proc.ProcessInput(context.Background(), input)
	if len(input.Tools) != 0 {
		t.Error("expected no tools with empty messages")
	}
}

func TestToolSearch_ProcessorNoUserMessage(t *testing.T) {
	ts := NewToolSearch(ToolSearchConfig{MinScore: 0.01})
	ts.AddTool(ToolDefinition{Name: "test", Description: "Test"}, "test")
	proc := ts.Processor()

	input := &StepInput{Messages: []Message{{Role: RoleAssistant, Content: "hello"}}}
	proc.ProcessInput(context.Background(), input)
	if len(input.Tools) != 0 {
		t.Error("expected no tools with no user message")
	}
}

func TestExtractKeywords(t *testing.T) {
	keywords := extractKeywords("Deploy the Node.js application to production")
	if len(keywords) == 0 {
		t.Fatal("expected keywords")
	}
	// Should include "deploy", "node", "application", "production"
	found := map[string]bool{}
	for _, kw := range keywords {
		found[kw] = true
	}
	if !found["deploy"] {
		t.Error("expected 'deploy' keyword")
	}
	if !found["production"] {
		t.Error("expected 'production' keyword")
	}
}

func TestExtractKeywords_FiltersStopWords(t *testing.T) {
	keywords := extractKeywords("the and for are but not")
	if len(keywords) != 0 {
		t.Errorf("expected 0 keywords from stop words, got %d: %v", len(keywords), keywords)
	}
}

func TestExtractKeywords_FiltersShortWords(t *testing.T) {
	keywords := extractKeywords("go is ok to do")
	if len(keywords) != 0 {
		t.Errorf("expected 0 keywords from short words, got %d: %v", len(keywords), keywords)
	}
}

func TestIsStopWord(t *testing.T) {
	if !isStopWord("the") {
		t.Error("expected 'the' to be stop word")
	}
	if isStopWord("deploy") {
		t.Error("expected 'deploy' to not be stop word")
	}
}

func TestBm25Score_NoMatch(t *testing.T) {
	score := bm25Score([]string{"quantum"}, []string{"deploy", "app"})
	if score != 0 {
		t.Errorf("expected 0 for no match, got %f", score)
	}
}

func TestBm25Score_EmptyDoc(t *testing.T) {
	score := bm25Score([]string{"test"}, []string{})
	if score != 0 {
		t.Errorf("expected 0 for empty doc, got %f", score)
	}
}

func TestBm25Score_Match(t *testing.T) {
	score := bm25Score([]string{"deploy", "app"}, []string{"deploy", "application", "app", "release"})
	if score <= 0 {
		t.Errorf("expected positive score, got %f", score)
	}
}
