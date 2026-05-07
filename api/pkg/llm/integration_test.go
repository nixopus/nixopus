//go:build integration

package llm

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

func getProvider(t *testing.T) Provider {
	t.Helper()

	apiKey := os.Getenv("LLM_API_KEY")
	if apiKey == "" {
		t.Skip("LLM_API_KEY not set, skipping integration test")
	}

	baseURL := os.Getenv("LLM_BASE_URL")
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	headers := map[string]string{}
	if referer := os.Getenv("LLM_HTTP_REFERER"); referer != "" {
		headers["HTTP-Referer"] = referer
	}

	return NewOpenAIProvider(OpenAIConfig{
		APIKey:  apiKey,
		BaseURL: baseURL,
		Headers: headers,
	})
}

func getModel() string {
	if m := os.Getenv("LLM_MODEL"); m != "" {
		return m
	}
	return "openai/gpt-4.1-mini"
}

func TestIntegration_Complete(t *testing.T) {
	provider := getProvider(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := provider.Complete(ctx, CompletionParams{
		Model: getModel(),
		Messages: []Message{
			{Role: RoleSystem, Content: "You are a helpful assistant. Be concise."},
			{Role: RoleUser, Content: "What is 2 + 2? Reply with just the number."},
		},
	})
	if err != nil {
		t.Fatalf("Complete failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("no choices returned")
	}

	t.Logf("Model: %s", resp.Model)
	t.Logf("Response: %s", resp.Choices[0].Message.Content)
	t.Logf("Usage: prompt=%d completion=%d total=%d",
		resp.Usage.PromptTokens, resp.Usage.CompletionTokens, resp.Usage.TotalTokens)

	if resp.Choices[0].Message.Content == "" {
		t.Error("empty response content")
	}
}

func TestIntegration_CompleteWithTools(t *testing.T) {
	provider := getProvider(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := provider.Complete(ctx, CompletionParams{
		Model: getModel(),
		Messages: []Message{
			{Role: RoleSystem, Content: "You have access to tools. Use them when appropriate."},
			{Role: RoleUser, Content: "What's the weather in Tokyo right now?"},
		},
		Tools: []Tool{
			{
				Type: "function",
				Function: ToolFunction{
					Name:        "get_weather",
					Description: "Get the current weather for a given city",
					Parameters: json.RawMessage(`{
						"type": "object",
						"properties": {
							"city": {"type": "string", "description": "City name"},
							"unit": {"type": "string", "enum": ["celsius", "fahrenheit"], "description": "Temperature unit"}
						},
						"required": ["city"]
					}`),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Complete with tools failed: %v", err)
	}

	if len(resp.Choices) == 0 {
		t.Fatal("no choices returned")
	}

	choice := resp.Choices[0]
	t.Logf("Finish reason: %s", choice.FinishReason)

	if choice.FinishReason == "tool_calls" {
		if len(choice.Message.ToolCalls) == 0 {
			t.Fatal("finish_reason is tool_calls but no tool calls present")
		}
		tc := choice.Message.ToolCalls[0]
		t.Logf("Tool call: %s(%s)", tc.Function.Name, tc.Function.Arguments)

		if tc.Function.Name != "get_weather" {
			t.Errorf("expected get_weather tool call, got %s", tc.Function.Name)
		}

		var args map[string]interface{}
		if err := json.Unmarshal([]byte(tc.Function.Arguments), &args); err != nil {
			t.Fatalf("invalid tool call arguments JSON: %v", err)
		}
		if _, ok := args["city"]; !ok {
			t.Error("tool call arguments missing 'city' field")
		}
	} else {
		t.Logf("Model responded directly: %s", choice.Message.Content)
	}
}

func TestIntegration_Stream(t *testing.T) {
	provider := getProvider(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	iter, err := provider.Stream(ctx, CompletionParams{
		Model: getModel(),
		Messages: []Message{
			{Role: RoleUser, Content: "Count from 1 to 5, separated by commas."},
		},
	})
	if err != nil {
		t.Fatalf("Stream failed: %v", err)
	}
	defer iter.Close()

	var content string
	var chunks int

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		switch event.Type {
		case EventChunk:
			chunks++
			if len(event.Chunk.Choices) > 0 {
				content += event.Chunk.Choices[0].Delta.Content
			}
		case EventDone:
			t.Log("Stream completed")
		case EventError:
			t.Fatalf("Stream error: %v", event.Err)
		}
	}

	t.Logf("Received %d chunks", chunks)
	t.Logf("Full content: %s", content)

	if content == "" {
		t.Error("no content received from stream")
	}
	if chunks < 2 {
		t.Errorf("expected multiple chunks, got %d", chunks)
	}
}

func TestIntegration_StreamWithToolCalls(t *testing.T) {
	provider := getProvider(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	iter, err := provider.Stream(ctx, CompletionParams{
		Model: getModel(),
		Messages: []Message{
			{Role: RoleUser, Content: "What's the weather in Paris?"},
		},
		Tools: []Tool{
			{
				Type: "function",
				Function: ToolFunction{
					Name:        "get_weather",
					Description: "Get the current weather for a given city",
					Parameters:  json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}`),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("Stream with tools failed: %v", err)
	}
	defer iter.Close()

	var toolName string
	var toolArgs string
	var chunks int

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Type == EventChunk {
			chunks++
			if len(event.Chunk.Choices) > 0 {
				delta := event.Chunk.Choices[0].Delta
				for _, tc := range delta.ToolCalls {
					if tc.Function.Name != "" {
						toolName = tc.Function.Name
					}
					toolArgs += tc.Function.Arguments
				}
			}
		}
		if event.Type == EventError {
			t.Fatalf("Stream error: %v", event.Err)
		}
	}

	t.Logf("Chunks: %d, Tool: %s, Args: %s", chunks, toolName, toolArgs)

	if toolName == "" && toolArgs == "" {
		t.Log("Model chose not to use tools (acceptable)")
		return
	}

	if toolName != "get_weather" {
		t.Errorf("expected get_weather, got %s", toolName)
	}
}
