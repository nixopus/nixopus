package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestComplete_BasicResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/chat/completions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected auth header: %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("unexpected content-type: %s", r.Header.Get("Content-Type"))
		}

		var params CompletionParams
		if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if params.Model != "gpt-4" {
			t.Errorf("expected model gpt-4, got %s", params.Model)
		}
		if len(params.Messages) != 1 || params.Messages[0].Content != "hello" {
			t.Errorf("unexpected messages: %+v", params.Messages)
		}

		resp := Response{
			ID:    "chatcmpl-123",
			Model: "gpt-4",
			Choices: []Choice{
				{
					Index:        0,
					Message:      Message{Role: RoleAssistant, Content: "Hello! How can I help?"},
					FinishReason: "stop",
				},
			},
			Usage: Usage{PromptTokens: 5, CompletionTokens: 7, TotalTokens: 12},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	resp, err := provider.Complete(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.ID != "chatcmpl-123" {
		t.Errorf("unexpected id: %s", resp.ID)
	}
	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	if resp.Choices[0].Message.Content != "Hello! How can I help?" {
		t.Errorf("unexpected content: %s", resp.Choices[0].Message.Content)
	}
	if resp.Usage.TotalTokens != 12 {
		t.Errorf("unexpected total tokens: %d", resp.Usage.TotalTokens)
	}
}

func TestComplete_ToolCalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var params CompletionParams
		json.NewDecoder(r.Body).Decode(&params)

		if len(params.Tools) != 1 {
			t.Fatalf("expected 1 tool, got %d", len(params.Tools))
		}
		if params.Tools[0].Function.Name != "get_weather" {
			t.Errorf("unexpected tool name: %s", params.Tools[0].Function.Name)
		}

		resp := Response{
			ID:    "chatcmpl-456",
			Model: "gpt-4",
			Choices: []Choice{
				{
					Index: 0,
					Message: Message{
						Role: RoleAssistant,
						ToolCalls: []ToolCall{
							{
								ID:   "call_abc123",
								Type: "function",
								Function: FunctionCall{
									Name:      "get_weather",
									Arguments: `{"city":"London"}`,
								},
							},
						},
					},
					FinishReason: "tool_calls",
				},
			},
			Usage: Usage{PromptTokens: 20, CompletionTokens: 15, TotalTokens: 35},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	resp, err := provider.Complete(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "what's the weather in London?"}},
		Tools: []Tool{
			{
				Type: "function",
				Function: ToolFunction{
					Name:        "get_weather",
					Description: "Get current weather for a city",
					Parameters:  json.RawMessage(`{"type":"object","properties":{"city":{"type":"string"}},"required":["city"]}`),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(resp.Choices) != 1 {
		t.Fatalf("expected 1 choice, got %d", len(resp.Choices))
	}
	choice := resp.Choices[0]
	if choice.FinishReason != "tool_calls" {
		t.Errorf("expected finish_reason tool_calls, got %s", choice.FinishReason)
	}
	if len(choice.Message.ToolCalls) != 1 {
		t.Fatalf("expected 1 tool call, got %d", len(choice.Message.ToolCalls))
	}
	tc := choice.Message.ToolCalls[0]
	if tc.ID != "call_abc123" {
		t.Errorf("unexpected tool call id: %s", tc.ID)
	}
	if tc.Function.Name != "get_weather" {
		t.Errorf("unexpected function name: %s", tc.Function.Name)
	}
	if tc.Function.Arguments != `{"city":"London"}` {
		t.Errorf("unexpected arguments: %s", tc.Function.Arguments)
	}
}

func TestComplete_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Rate limit exceeded",
				"type":    "rate_limit_error",
				"code":    "rate_limit_exceeded",
			},
		})
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	_, err := provider.Complete(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	provErr, ok := err.(*ProviderError)
	if !ok {
		t.Fatalf("expected *ProviderError, got %T", err)
	}
	if provErr.StatusCode != 429 {
		t.Errorf("expected status 429, got %d", provErr.StatusCode)
	}
	if provErr.Type != "rate_limit_error" {
		t.Errorf("unexpected error type: %s", provErr.Type)
	}
}

func TestComplete_CustomHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("HTTP-Referer") != "https://nixopus.com" {
			t.Errorf("missing custom header HTTP-Referer")
		}
		if r.Header.Get("X-Title") != "Nixopus Agent" {
			t.Errorf("missing custom header X-Title")
		}

		resp := Response{
			ID:    "chatcmpl-789",
			Model: "openrouter/anthropic/claude-sonnet-4-20250514",
			Choices: []Choice{
				{Index: 0, Message: Message{Role: RoleAssistant, Content: "ok"}, FinishReason: "stop"},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "or-key",
		BaseURL: server.URL,
		Headers: map[string]string{
			"HTTP-Referer": "https://nixopus.com",
			"X-Title":      "Nixopus Agent",
		},
	})

	_, err := provider.Complete(context.Background(), CompletionParams{
		Model:    "openrouter/anthropic/claude-sonnet-4-20250514",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStream_BasicResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["stream"] != true {
			t.Error("expected stream: true in request")
		}

		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)

		chunks := []string{
			`{"id":"chatcmpl-stream","model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}`,
			`{"id":"chatcmpl-stream","model":"gpt-4","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}`,
			`{"id":"chatcmpl-stream","model":"gpt-4","choices":[{"index":0,"delta":{"content":" world"},"finish_reason":null}]}`,
			`{"id":"chatcmpl-stream","model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":2,"total_tokens":7}}`,
		}

		for _, chunk := range chunks {
			fmt.Fprintf(w, "data: %s\n\n", chunk)
			flusher.Flush()
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	iter, err := provider.Stream(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hello"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer iter.Close()

	var content string
	var gotDone bool
	var lastUsage *Usage

	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		switch event.Type {
		case EventChunk:
			if len(event.Chunk.Choices) > 0 {
				content += event.Chunk.Choices[0].Delta.Content
			}
			if event.Chunk.Usage != nil {
				lastUsage = event.Chunk.Usage
			}
		case EventDone:
			gotDone = true
		case EventError:
			t.Fatalf("unexpected stream error: %v", event.Err)
		}
	}

	if content != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", content)
	}
	if !gotDone {
		t.Error("did not receive done event")
	}
	if lastUsage == nil || lastUsage.TotalTokens != 7 {
		t.Errorf("unexpected usage: %+v", lastUsage)
	}
}

func TestStream_ToolCallChunks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)

		chunks := []string{
			`{"id":"chatcmpl-tc","model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"id":"call_1","type":"function","function":{"name":"get_weather","arguments":""}}]},"finish_reason":null}]}`,
			`{"id":"chatcmpl-tc","model":"gpt-4","choices":[{"index":0,"delta":{"tool_calls":[{"id":"","type":"","function":{"name":"","arguments":"{\"city\":"}}]},"finish_reason":null}]}`,
			`{"id":"chatcmpl-tc","model":"gpt-4","choices":[{"index":0,"delta":{"tool_calls":[{"id":"","type":"","function":{"name":"","arguments":"\"London\"}"}}]},"finish_reason":null}]}`,
			`{"id":"chatcmpl-tc","model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}`,
		}

		for _, chunk := range chunks {
			fmt.Fprintf(w, "data: %s\n\n", chunk)
			flusher.Flush()
		}
		fmt.Fprint(w, "data: [DONE]\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	iter, err := provider.Stream(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "weather?"}},
		Tools: []Tool{{
			Type:     "function",
			Function: ToolFunction{Name: "get_weather", Description: "Get weather", Parameters: json.RawMessage(`{}`)},
		}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer iter.Close()

	var toolCallChunks int
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Type == EventChunk && len(event.Chunk.Choices) > 0 {
			if len(event.Chunk.Choices[0].Delta.ToolCalls) > 0 {
				toolCallChunks++
			}
		}
	}

	if toolCallChunks != 3 {
		t.Errorf("expected 3 tool call chunks, got %d", toolCallChunks)
	}
}

func TestComplete_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{
		APIKey:  "test-key",
		BaseURL: server.URL,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := provider.Complete(ctx, CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestNewOpenAIProvider_DefaultBaseURL(t *testing.T) {
	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key"})
	p := provider.(*openaiProvider)
	if p.config.BaseURL != "https://api.openai.com/v1" {
		t.Errorf("unexpected default base url: %s", p.config.BaseURL)
	}
}

func TestNewOpenAIProvider_TrailingSlashTrimmed(t *testing.T) {
	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: "https://openrouter.ai/api/v1/"})
	p := provider.(*openaiProvider)
	if p.config.BaseURL != "https://openrouter.ai/api/v1" {
		t.Errorf("trailing slash not trimmed: %s", p.config.BaseURL)
	}
}

func TestProviderError_ErrorWithType(t *testing.T) {
	err := &ProviderError{
		StatusCode: 429,
		Type:       "rate_limit_error",
		Code:       "rate_limit_exceeded",
		Message:    "Too many requests",
	}
	expected := "llm: api error 429 [rate_limit_error/rate_limit_exceeded]: Too many requests"
	if err.Error() != expected {
		t.Errorf("got %q, want %q", err.Error(), expected)
	}
}

func TestProviderError_ErrorWithoutType(t *testing.T) {
	err := &ProviderError{
		StatusCode: 500,
		Message:    "Internal server error",
	}
	expected := "llm: api error 500: Internal server error"
	if err.Error() != expected {
		t.Errorf("got %q, want %q", err.Error(), expected)
	}
}

func TestComplete_InvalidResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	_, err := provider.Complete(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for invalid response body")
	}
	if !contains(err.Error(), "decode response") {
		t.Errorf("expected decode error, got: %v", err)
	}
}

func TestComplete_InvalidBaseURL(t *testing.T) {
	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: "http://[::1]:namedport"})
	_, err := provider.Complete(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for invalid base URL")
	}
}

func TestStream_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":{"message":"Service unavailable","type":"server_error","code":"unavailable"}}`))
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	_, err := provider.Stream(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for 503 response")
	}
	provErr, ok := err.(*ProviderError)
	if !ok {
		t.Fatalf("expected *ProviderError, got %T", err)
	}
	if provErr.StatusCode != 503 {
		t.Errorf("expected 503, got %d", provErr.StatusCode)
	}
}

func TestStream_InvalidBaseURL(t *testing.T) {
	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: "http://[::1]:namedport"})
	_, err := provider.Stream(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error for invalid base URL")
	}
}

func TestStream_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		fmt.Fprintf(w, "data: %s\n\n", `{"id":"1","model":"gpt-4","choices":[{"index":0,"delta":{"content":"hello"},"finish_reason":null}]}`)
		flusher.Flush()
		<-r.Context().Done()
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	ctx, cancel := context.WithCancel(context.Background())

	iter, err := provider.Stream(ctx, CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	event, ok := iter.Next()
	if !ok || event.Type != EventChunk {
		t.Fatal("expected first chunk")
	}

	cancel()
	iter.Close()
}

func TestStream_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		fmt.Fprintf(w, "data: {invalid json}\n\n")
		flusher.Flush()
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	iter, err := provider.Stream(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer iter.Close()

	event, ok := iter.Next()
	if !ok {
		t.Fatal("expected an event")
	}
	if event.Type != EventError {
		t.Errorf("expected EventError, got %d", event.Type)
	}
	if !contains(event.Err.Error(), "decode chunk") {
		t.Errorf("expected decode chunk error, got: %v", event.Err)
	}
}

func TestParseError_NonJSONBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`<html>502 Bad Gateway</html>`))
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	_, err := provider.Complete(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	provErr, ok := err.(*ProviderError)
	if !ok {
		t.Fatalf("expected *ProviderError, got %T", err)
	}
	if provErr.StatusCode != 502 {
		t.Errorf("expected 502, got %d", provErr.StatusCode)
	}
	if provErr.Type != "" {
		t.Errorf("expected empty type for non-JSON, got %s", provErr.Type)
	}
	if !contains(provErr.Message, "502 Bad Gateway") {
		t.Errorf("expected raw body in message, got: %s", provErr.Message)
	}
}

func TestStream_RequestFailure(t *testing.T) {
	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: "http://127.0.0.1:1"})
	_, err := provider.Stream(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected connection error")
	}
	if !contains(err.Error(), "stream request failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestComplete_RequestFailure(t *testing.T) {
	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: "http://127.0.0.1:1"})
	_, err := provider.Complete(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err == nil {
		t.Fatal("expected connection error")
	}
	if !contains(err.Error(), "request failed") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestComplete_MarshalError(t *testing.T) {
	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: "http://localhost"})

	nan := math.NaN()
	_, err := provider.Complete(context.Background(), CompletionParams{
		Model:       "gpt-4",
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Temperature: &nan,
	})
	if err == nil {
		t.Fatal("expected marshal error for NaN temperature")
	}
	if !strings.Contains(err.Error(), "marshal request") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStream_MarshalError(t *testing.T) {
	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: "http://localhost"})

	nan := math.NaN()
	_, err := provider.Stream(context.Background(), CompletionParams{
		Model:       "gpt-4",
		Messages:    []Message{{Role: RoleUser, Content: "hi"}},
		Temperature: &nan,
	})
	if err == nil {
		t.Fatal("expected marshal error for NaN temperature")
	}
	if !strings.Contains(err.Error(), "marshal stream request") {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStream_ScannerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		hijacker, ok := w.(http.Hijacker)
		if !ok {
			t.Fatal("server does not support hijacking")
		}
		conn, buf, err := hijacker.Hijack()
		if err != nil {
			t.Fatalf("hijack failed: %v", err)
		}
		buf.WriteString("HTTP/1.1 200 OK\r\nContent-Type: text/event-stream\r\n\r\n")
		buf.WriteString("data: {\"id\":\"1\",\"model\":\"gpt-4\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"hi\"},\"finish_reason\":null}]}\n\n")
		buf.Flush()
		conn.Close()
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	iter, err := provider.Stream(context.Background(), CompletionParams{
		Model:    "gpt-4",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer iter.Close()

	var gotChunk, gotEnd bool
	for {
		event, ok := iter.Next()
		if !ok {
			gotEnd = true
			break
		}
		if event.Type == EventChunk {
			gotChunk = true
		}
		if event.Type == EventError || event.Type == EventDone {
			gotEnd = true
			break
		}
	}

	if !gotChunk {
		t.Error("expected at least one chunk before connection close")
	}
	if !gotEnd {
		t.Error("expected stream to end")
	}
}

func TestReadSSE_ReaderError(t *testing.T) {
	p := &openaiProvider{}
	ch := make(chan StreamEvent, 16)

	errReader := &failingReader{
		data: []byte("data: {\"id\":\"1\",\"model\":\"gpt-4\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"hi\"},\"finish_reason\":null}]}\n\n"),
		err:  fmt.Errorf("connection reset"),
	}

	ctx := context.Background()
	go p.readSSE(ctx, io.NopCloser(errReader), ch)

	var gotChunk, gotError bool
	for event := range ch {
		switch event.Type {
		case EventChunk:
			gotChunk = true
		case EventError:
			gotError = true
			if !strings.Contains(event.Err.Error(), "read stream") {
				t.Errorf("expected 'read stream' error, got: %v", event.Err)
			}
		}
	}

	if !gotChunk {
		t.Error("expected a chunk before error")
	}
	if !gotError {
		t.Error("expected error event from reader failure")
	}
}

type failingReader struct {
	data []byte
	pos  int
	err  error
}

func (r *failingReader) Read(p []byte) (int, error) {
	if r.pos < len(r.data) {
		n := copy(p, r.data[r.pos:])
		r.pos += n
		return n, nil
	}
	return 0, r.err
}

func TestReadSSE_ContextCancelledDuringLoop(t *testing.T) {
	p := &openaiProvider{}
	ch := make(chan StreamEvent, 16)

	data := "data: {\"id\":\"1\",\"model\":\"gpt-4\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"a\"},\"finish_reason\":null}]}\n\ndata: {\"id\":\"1\",\"model\":\"gpt-4\",\"choices\":[{\"index\":0,\"delta\":{\"content\":\"b\"},\"finish_reason\":null}]}\n\n"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	p.readSSE(ctx, io.NopCloser(strings.NewReader(data)), ch)

	var events []StreamEvent
	for e := range ch {
		events = append(events, e)
	}
	if len(events) != 0 {
		t.Errorf("expected no events when context already cancelled, got %d", len(events))
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
