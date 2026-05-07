package llm

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func streamingServer(responses ...string) *httptest.Server {
	var callCount atomic.Int32
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := int(callCount.Add(1)) - 1
		if idx >= len(responses) {
			idx = len(responses) - 1
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)

		if body["stream"] == true {
			w.Header().Set("Content-Type", "text/event-stream")
			flusher := w.(http.Flusher)
			fmt.Fprint(w, responses[idx])
			flusher.Flush()
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(responses[idx]))
		}
	}))
}

func makeSSEResponse(chunks []string) string {
	var sb strings.Builder
	for _, c := range chunks {
		sb.WriteString("data: ")
		sb.WriteString(c)
		sb.WriteString("\n\n")
	}
	sb.WriteString("data: [DONE]\n\n")
	return sb.String()
}

func parseSSEEvents(body string) []SSEEvent {
	var events []SSEEvent
	scanner := bufio.NewScanner(strings.NewReader(body))
	var currentEvent, currentData string

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
		} else if strings.HasPrefix(line, "data: ") {
			currentData = strings.TrimPrefix(line, "data: ")
		} else if line == "" && currentEvent != "" {
			events = append(events, SSEEvent{Event: currentEvent, Data: currentData})
			currentEvent = ""
			currentData = ""
		}
	}
	return events
}

func TestStreamHandler_BasicCompletion(t *testing.T) {
	sseResp := makeSSEResponse([]string{
		`{"id":"1","model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","content":""},"finish_reason":null}]}`,
		`{"id":"1","model":"gpt-4","choices":[{"index":0,"delta":{"content":"Hello"},"finish_reason":null}]}`,
		`{"id":"1","model":"gpt-4","choices":[{"index":0,"delta":{"content":" world"},"finish_reason":null}]}`,
		`{"id":"1","model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":5,"completion_tokens":2,"total_tokens":7}}`,
	})

	server := streamingServer(sseResp)
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	tools := NewToolRegistry()
	agent := NewAgent(provider, tools, AgentConfig{Model: "gpt-4", SystemPrompt: "Be helpful"})
	handler := NewStreamHandler(agent)

	req := httptest.NewRequest("POST", "/chat", strings.NewReader(`{"input":"hi"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	events := parseSSEEvents(rec.Body.String())
	var contentParts []string
	var gotDone bool

	for _, e := range events {
		switch e.Event {
		case "content":
			var s string
			if err := json.Unmarshal([]byte(e.Data), &s); err != nil {
				contentParts = append(contentParts, e.Data)
			} else {
				contentParts = append(contentParts, s)
			}
		case "done":
			gotDone = true
		}
	}

	fullContent := strings.Join(contentParts, "")
	if fullContent != "Hello world" {
		t.Errorf("expected 'Hello world', got %q", fullContent)
	}
	if !gotDone {
		t.Error("expected done event")
	}
}

func TestStreamHandler_WithToolCalls(t *testing.T) {
	// First response: tool call
	toolCallSSE := makeSSEResponse([]string{
		`{"id":"1","model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"id":"call_1","type":"function","function":{"name":"get_weather","arguments":""}}]},"finish_reason":null}]}`,
		`{"id":"1","model":"gpt-4","choices":[{"index":0,"delta":{"tool_calls":[{"id":"","type":"","function":{"name":"","arguments":"{\"city\":\"London\"}"}}]},"finish_reason":null}]}`,
		`{"id":"1","model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}],"usage":{"prompt_tokens":20,"completion_tokens":10,"total_tokens":30}}`,
	})
	// Second response: final answer
	finalSSE := makeSSEResponse([]string{
		`{"id":"2","model":"gpt-4","choices":[{"index":0,"delta":{"content":"It's 18°C."},"finish_reason":null}]}`,
		`{"id":"2","model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}],"usage":{"prompt_tokens":40,"completion_tokens":5,"total_tokens":45}}`,
	})

	server := streamingServer(toolCallSSE, finalSSE)
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	tools := NewToolRegistry()
	tools.Register(ToolDefinition{
		Name: "get_weather",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`{"temp":18}`), nil
		},
	})

	agent := NewAgent(provider, tools, AgentConfig{Model: "gpt-4"})
	handler := NewStreamHandler(agent)

	req := httptest.NewRequest("POST", "/chat", strings.NewReader(`{"input":"weather in London?"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	events := parseSSEEvents(rec.Body.String())
	var gotToolCalls, gotToolResult, gotContent, gotDone bool

	for _, e := range events {
		switch e.Event {
		case "tool_calls":
			gotToolCalls = true
		case "tool_result":
			gotToolResult = true
			if !strings.Contains(e.Data, "temp") {
				t.Errorf("unexpected tool result: %s", e.Data)
			}
		case "content":
			gotContent = true
		case "done":
			gotDone = true
		}
	}

	if !gotToolCalls {
		t.Error("expected tool_calls event")
	}
	if !gotToolResult {
		t.Error("expected tool_result event")
	}
	if !gotContent {
		t.Error("expected content event")
	}
	if !gotDone {
		t.Error("expected done event")
	}
}

func TestStreamHandler_InvalidBody(t *testing.T) {
	agent := NewAgent(nil, NewToolRegistry(), AgentConfig{Model: "gpt-4"})
	handler := NewStreamHandler(agent)

	req := httptest.NewRequest("POST", "/chat", strings.NewReader(`invalid json`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestStreamHandler_EmptyInput(t *testing.T) {
	agent := NewAgent(nil, NewToolRegistry(), AgentConfig{Model: "gpt-4"})
	handler := NewStreamHandler(agent)

	req := httptest.NewRequest("POST", "/chat", strings.NewReader(`{"input":""}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestStreamHandler_LLMError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]string{"message": "model overloaded"},
		})
	}))
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	agent := NewAgent(provider, NewToolRegistry(), AgentConfig{Model: "gpt-4"})
	handler := NewStreamHandler(agent)

	req := httptest.NewRequest("POST", "/chat", strings.NewReader(`{"input":"hi"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	events := parseSSEEvents(rec.Body.String())
	var gotError bool
	for _, e := range events {
		if e.Event == "error" {
			gotError = true
		}
	}
	if !gotError {
		t.Error("expected error event")
	}
}

func TestStreamHandler_MaxStepsExceeded(t *testing.T) {
	toolCallSSE := makeSSEResponse([]string{
		`{"id":"1","model":"gpt-4","choices":[{"index":0,"delta":{"role":"assistant","tool_calls":[{"id":"call_1","type":"function","function":{"name":"loop","arguments":"{}"}}]},"finish_reason":null}]}`,
		`{"id":"1","model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"tool_calls"}]}`,
	})

	server := streamingServer(toolCallSSE, toolCallSSE, toolCallSSE, toolCallSSE)
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	tools := NewToolRegistry()
	tools.Register(ToolDefinition{
		Name: "loop",
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			return json.RawMessage(`"again"`), nil
		},
	})

	agent := NewAgent(provider, tools, AgentConfig{Model: "gpt-4", MaxSteps: 2})
	handler := NewStreamHandler(agent)

	req := httptest.NewRequest("POST", "/chat", strings.NewReader(`{"input":"loop"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	events := parseSSEEvents(rec.Body.String())
	var gotError bool
	for _, e := range events {
		if e.Event == "error" && strings.Contains(e.Data, "max steps") {
			gotError = true
		}
	}
	if !gotError {
		t.Error("expected max steps error event")
	}
}

func TestStreamHandler_StreamError(t *testing.T) {
	brokenSSE := "data: {invalid json}\n\n"
	server := streamingServer(brokenSSE)
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	agent := NewAgent(provider, NewToolRegistry(), AgentConfig{Model: "gpt-4"})
	handler := NewStreamHandler(agent)

	req := httptest.NewRequest("POST", "/chat", strings.NewReader(`{"input":"hi"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	events := parseSSEEvents(rec.Body.String())
	var gotError bool
	for _, e := range events {
		if e.Event == "error" {
			gotError = true
		}
	}
	if !gotError {
		t.Error("expected error event from broken stream")
	}
}

type noFlushResponseWriter struct {
	code int
	body strings.Builder
	h    http.Header
}

func (w *noFlushResponseWriter) Header() http.Header {
	if w.h == nil {
		w.h = make(http.Header)
	}
	return w.h
}
func (w *noFlushResponseWriter) Write(b []byte) (int, error) { return w.body.Write(b) }
func (w *noFlushResponseWriter) WriteHeader(code int)        { w.code = code }

func TestStreamHandler_NoFlusher(t *testing.T) {
	agent := NewAgent(nil, NewToolRegistry(), AgentConfig{Model: "gpt-4"})
	handler := NewStreamHandler(agent)

	req := httptest.NewRequest("POST", "/chat", strings.NewReader(`{"input":"hi"}`))
	w := &noFlushResponseWriter{}
	handler.ServeHTTP(w, req)

	if w.code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.code)
	}
	if !strings.Contains(w.body.String(), "streaming not supported") {
		t.Errorf("unexpected body: %s", w.body.String())
	}
}

func TestStreamHandler_EmptyChoices(t *testing.T) {
	sseResp := makeSSEResponse([]string{
		`{"id":"1","model":"gpt-4","choices":[]}`,
		`{"id":"1","model":"gpt-4","choices":[{"index":0,"delta":{"content":"ok"},"finish_reason":null}]}`,
		`{"id":"1","model":"gpt-4","choices":[{"index":0,"delta":{},"finish_reason":"stop"}]}`,
	})

	server := streamingServer(sseResp)
	defer server.Close()

	provider := NewOpenAIProvider(OpenAIConfig{APIKey: "key", BaseURL: server.URL})
	agent := NewAgent(provider, NewToolRegistry(), AgentConfig{Model: "gpt-4"})
	handler := NewStreamHandler(agent)

	req := httptest.NewRequest("POST", "/chat", strings.NewReader(`{"input":"hi"}`))
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	events := parseSSEEvents(rec.Body.String())
	var gotDone bool
	for _, e := range events {
		if e.Event == "done" {
			gotDone = true
		}
	}
	if !gotDone {
		t.Error("expected done event")
	}
}
