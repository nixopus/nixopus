package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func TestAgent_SimpleCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			ID:    "1",
			Model: "gpt-4",
			Choices: []Choice{{
				Index:        0,
				Message:      Message{Role: RoleAssistant, Content: "Hello!"},
				FinishReason: "stop",
			}},
			Usage: Usage{PromptTokens: 10, CompletionTokens: 5, TotalTokens: 15},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	tools := NewToolRegistry()
	agent := NewAgent(provider, tools, AgentConfig{
		Model:        "gpt-4",
		SystemPrompt: "You are helpful.",
	})

	result, err := agent.Run(context.Background(), "hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "Hello!" {
		t.Errorf("expected 'Hello!', got %q", result.Content)
	}
	if result.Steps != 1 {
		t.Errorf("expected 1 step, got %d", result.Steps)
	}
	if result.TotalUsage.TotalTokens != 15 {
		t.Errorf("expected 15 total tokens, got %d", result.TotalUsage.TotalTokens)
	}
	if len(result.Messages) != 3 {
		t.Errorf("expected 3 messages (system+user+assistant), got %d", len(result.Messages))
	}
}

func TestAgent_WithToolCalls(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)

		var resp Response
		if count == 1 {
			resp = Response{
				ID:    "1",
				Model: "gpt-4",
				Choices: []Choice{{
					Index: 0,
					Message: Message{
						Role: RoleAssistant,
						ToolCalls: []ToolCall{{
							ID:       "call_1",
							Type:     "function",
							Function: FunctionCall{Name: "get_weather", Arguments: `{"city":"London"}`},
						}},
					},
					FinishReason: "tool_calls",
				}},
				Usage: Usage{PromptTokens: 20, CompletionTokens: 10, TotalTokens: 30},
			}
		} else {
			resp = Response{
				ID:    "2",
				Model: "gpt-4",
				Choices: []Choice{{
					Index:        0,
					Message:      Message{Role: RoleAssistant, Content: "It's 18°C in London."},
					FinishReason: "stop",
				}},
				Usage: Usage{PromptTokens: 40, CompletionTokens: 10, TotalTokens: 50},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	tools := NewToolRegistry()
	tools.Register(ToolDefinition{
		Name:        "get_weather",
		Description: "Get weather",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				City string `json:"city"`
			}
			json.Unmarshal(args, &input)
			if input.City != "London" {
				t.Errorf("expected London, got %s", input.City)
			}
			return json.RawMessage(`{"temp":18,"unit":"celsius"}`), nil
		},
	})

	agent := NewAgent(provider, tools, AgentConfig{
		Model:        "gpt-4",
		SystemPrompt: "You can check weather.",
	})

	result, err := agent.Run(context.Background(), "What's the weather in London?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "It's 18°C in London." {
		t.Errorf("unexpected content: %s", result.Content)
	}
	if result.Steps != 2 {
		t.Errorf("expected 2 steps, got %d", result.Steps)
	}
	if result.TotalUsage.TotalTokens != 80 {
		t.Errorf("expected 80 total tokens, got %d", result.TotalUsage.TotalTokens)
	}
}

func TestAgent_MultipleToolCalls(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)

		var resp Response
		if count == 1 {
			resp = Response{
				ID:    "1",
				Model: "gpt-4",
				Choices: []Choice{{
					Index: 0,
					Message: Message{
						Role: RoleAssistant,
						ToolCalls: []ToolCall{
							{ID: "call_1", Type: "function", Function: FunctionCall{Name: "get_weather", Arguments: `{"city":"London"}`}},
							{ID: "call_2", Type: "function", Function: FunctionCall{Name: "get_weather", Arguments: `{"city":"Paris"}`}},
						},
					},
					FinishReason: "tool_calls",
				}},
				Usage: Usage{PromptTokens: 20, CompletionTokens: 15, TotalTokens: 35},
			}
		} else {
			resp = Response{
				ID:    "2",
				Model: "gpt-4",
				Choices: []Choice{{
					Index:        0,
					Message:      Message{Role: RoleAssistant, Content: "London: 18°C, Paris: 22°C"},
					FinishReason: "stop",
				}},
				Usage: Usage{PromptTokens: 60, CompletionTokens: 12, TotalTokens: 72},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	tools := NewToolRegistry()
	tools.Register(ToolDefinition{
		Name:       "get_weather",
		Parameters: json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}}}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				City string `json:"city"`
			}
			json.Unmarshal(args, &input)
			temps := map[string]int{"London": 18, "Paris": 22}
			return json.Marshal(map[string]int{"temp": temps[input.City]})
		},
	})

	agent := NewAgent(provider, tools, AgentConfig{Model: "gpt-4"})
	result, err := agent.Run(context.Background(), "Weather in London and Paris?")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Steps != 2 {
		t.Errorf("expected 2 steps, got %d", result.Steps)
	}
	// system(none) + user + assistant(tool_calls) + tool(London) + tool(Paris) + assistant(final)
	if len(result.Messages) != 5 {
		t.Errorf("expected 5 messages, got %d", len(result.Messages))
	}
}

func TestAgent_ToolError(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)

		var resp Response
		if count == 1 {
			resp = Response{
				ID:    "1",
				Model: "gpt-4",
				Choices: []Choice{{
					Index: 0,
					Message: Message{
						Role:      RoleAssistant,
						ToolCalls: []ToolCall{{ID: "call_1", Type: "function", Function: FunctionCall{Name: "fail_tool", Arguments: `{}`}}},
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
					Index:        0,
					Message:      Message{Role: RoleAssistant, Content: "The tool failed, sorry."},
					FinishReason: "stop",
				}},
				Usage: Usage{TotalTokens: 30},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	tools := NewToolRegistry()
	tools.Register(ToolDefinition{
		Name: "fail_tool",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return nil, fmt.Errorf("database connection timeout")
		},
	})

	agent := NewAgent(provider, tools, AgentConfig{Model: "gpt-4"})
	result, err := agent.Run(context.Background(), "do the thing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Content != "The tool failed, sorry." {
		t.Errorf("unexpected content: %s", result.Content)
	}
	// Verify the error message was sent back to the model
	toolMsg := result.Messages[2] // user + assistant(tool_calls) + tool_result
	if toolMsg.Role != RoleTool {
		t.Fatalf("expected tool role, got %s", toolMsg.Role)
	}
	if !strings.Contains(toolMsg.Content, "database connection timeout") {
		t.Errorf("expected error in tool message, got: %s", toolMsg.Content)
	}
}

func TestAgent_MaxStepsExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{
			ID:    "1",
			Model: "gpt-4",
			Choices: []Choice{{
				Index: 0,
				Message: Message{
					Role:      RoleAssistant,
					ToolCalls: []ToolCall{{ID: "call_1", Type: "function", Function: FunctionCall{Name: "loop", Arguments: `{}`}}},
				},
				FinishReason: "tool_calls",
			}},
			Usage: Usage{TotalTokens: 10},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	tools := NewToolRegistry()
	tools.Register(ToolDefinition{
		Name: "loop",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`"again"`), nil
		},
	})

	agent := NewAgent(provider, tools, AgentConfig{Model: "gpt-4", MaxSteps: 3})
	_, err := agent.Run(context.Background(), "loop forever")
	if err == nil {
		t.Fatal("expected max steps error")
	}
	if !strings.Contains(err.Error(), "exceeded max steps") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAgent_LLMError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "model overloaded", "type": "server_error"},
		})
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	tools := NewToolRegistry()
	agent := NewAgent(provider, tools, AgentConfig{Model: "gpt-4"})

	_, err := agent.Run(context.Background(), "hi")
	if err == nil {
		t.Fatal("expected error from LLM")
	}
	if !strings.Contains(err.Error(), "agent step 1 failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAgent_EmptyChoices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Response{ID: "1", Model: "gpt-4", Choices: []Choice{}}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	tools := NewToolRegistry()
	agent := NewAgent(provider, tools, AgentConfig{Model: "gpt-4"})

	_, err := agent.Run(context.Background(), "hi")
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
	if !strings.Contains(err.Error(), "no choices") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestAgent_WithHistory(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var params CompletionParams
		json.NewDecoder(r.Body).Decode(&params)

		// Verify message structure: system + history(2) + user
		if len(params.Messages) != 4 {
			t.Errorf("expected 4 messages, got %d", len(params.Messages))
		}
		if params.Messages[0].Role != RoleSystem {
			t.Errorf("expected system first, got %s", params.Messages[0].Role)
		}
		if params.Messages[1].Role != RoleUser || params.Messages[1].Content != "previous question" {
			t.Errorf("unexpected history[0]: %+v", params.Messages[1])
		}
		if params.Messages[2].Role != RoleAssistant || params.Messages[2].Content != "previous answer" {
			t.Errorf("unexpected history[1]: %+v", params.Messages[2])
		}
		if params.Messages[3].Role != RoleUser || params.Messages[3].Content != "follow up" {
			t.Errorf("unexpected last message: %+v", params.Messages[3])
		}

		resp := Response{
			ID:    "1",
			Model: "gpt-4",
			Choices: []Choice{{
				Index:        0,
				Message:      Message{Role: RoleAssistant, Content: "follow up answer"},
				FinishReason: "stop",
			}},
			Usage: Usage{TotalTokens: 20},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	agent := NewAgent(provider, NewToolRegistry(), AgentConfig{
		Model:        "gpt-4",
		SystemPrompt: "You are helpful.",
	})

	history := []Message{
		{Role: RoleUser, Content: "previous question"},
		{Role: RoleAssistant, Content: "previous answer"},
	}

	result, err := agent.Run(context.Background(), "follow up", history...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "follow up answer" {
		t.Errorf("unexpected content: %s", result.Content)
	}
}

func TestAgent_NoSystemPrompt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var params CompletionParams
		json.NewDecoder(r.Body).Decode(&params)

		if len(params.Messages) != 1 {
			t.Errorf("expected 1 message (user only), got %d", len(params.Messages))
		}
		if params.Messages[0].Role != RoleUser {
			t.Errorf("expected user role, got %s", params.Messages[0].Role)
		}

		resp := Response{
			ID:      "1",
			Model:   "gpt-4",
			Choices: []Choice{{Message: Message{Role: RoleAssistant, Content: "ok"}, FinishReason: "stop"}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	agent := NewAgent(provider, NewToolRegistry(), AgentConfig{Model: "gpt-4"})

	result, err := agent.Run(context.Background(), "hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "ok" {
		t.Errorf("unexpected content: %s", result.Content)
	}
}

func TestAgent_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	agent := NewAgent(provider, NewToolRegistry(), AgentConfig{Model: "gpt-4"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := agent.Run(ctx, "hi")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestAgent_NoToolsInParams(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var params CompletionParams
		json.NewDecoder(r.Body).Decode(&params)

		if len(params.Tools) != 0 {
			t.Errorf("expected no tools, got %d", len(params.Tools))
		}

		resp := Response{
			ID:      "1",
			Model:   "gpt-4",
			Choices: []Choice{{Message: Message{Role: RoleAssistant, Content: "done"}, FinishReason: "stop"}},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	agent := NewAgent(provider, NewToolRegistry(), AgentConfig{Model: "gpt-4"})

	_, err := agent.Run(context.Background(), "hi")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAgent_DefaultMaxSteps(t *testing.T) {
	agent := NewAgent(nil, NewToolRegistry(), AgentConfig{Model: "gpt-4"})
	if agent.config.MaxSteps != 25 {
		t.Errorf("expected default max steps 25, got %d", agent.config.MaxSteps)
	}
}

func TestAgent_UnknownToolCall(t *testing.T) {
	var callCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := callCount.Add(1)

		var resp Response
		if count == 1 {
			resp = Response{
				ID:    "1",
				Model: "gpt-4",
				Choices: []Choice{{
					Index: 0,
					Message: Message{
						Role:      RoleAssistant,
						ToolCalls: []ToolCall{{ID: "call_1", Type: "function", Function: FunctionCall{Name: "nonexistent", Arguments: `{}`}}},
					},
					FinishReason: "tool_calls",
				}},
				Usage: Usage{TotalTokens: 10},
			}
		} else {
			resp = Response{
				ID:      "2",
				Model:   "gpt-4",
				Choices: []Choice{{Message: Message{Role: RoleAssistant, Content: "tool not found"}, FinishReason: "stop"}},
				Usage:   Usage{TotalTokens: 15},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	agent := NewAgent(provider, NewToolRegistry(), AgentConfig{Model: "gpt-4"})

	result, err := agent.Run(context.Background(), "use a tool")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// The agent should gracefully handle unknown tools by sending error back to model
	toolMsg := result.Messages[2]
	if toolMsg.Role != RoleTool {
		t.Fatalf("expected tool message, got %s", toolMsg.Role)
	}
	if !strings.Contains(toolMsg.Content, "not found") {
		t.Errorf("expected 'not found' in tool error, got: %s", toolMsg.Content)
	}
}
