package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// ErrStreamCancelled is returned when the client disconnects mid-stream.
var ErrStreamCancelled = errors.New("llm: stream cancelled by client")

type StreamHandler struct {
	agent    *Agent
	ThreadID string
	Model    string // per-request model override; empty = use agent default
	OnDone   func(result *StreamRunResult)
}

func NewStreamHandler(agent *Agent) *StreamHandler {
	return &StreamHandler{agent: agent}
}

type SSEEvent struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

type StreamRunResult struct {
	Content    string
	Messages   []Message
	TotalUsage Usage
	Steps      int
}

func (h *StreamHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	var req struct {
		Input   string    `json:"input"`
		History []Message `json:"history,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Input == "" {
		http.Error(w, "input is required", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ctx := r.Context()
	result, err := h.runStreaming(ctx, w, flusher, req.Input, req.History)

	if errors.Is(err, ErrStreamCancelled) || errors.Is(err, context.Canceled) {
		if result != nil && h.OnDone != nil && result.Content != "" {
			h.OnDone(result)
		}
		h.writeSSE(w, flusher, "cancelled", map[string]interface{}{
			"content":   result.Content,
			"usage":     result.TotalUsage,
			"steps":     result.Steps,
			"thread_id": h.ThreadID,
		})
		return
	}

	if err != nil {
		h.writeSSE(w, flusher, "error", map[string]string{"message": err.Error()})
		return
	}

	if h.OnDone != nil {
		h.OnDone(result)
	}

	h.writeSSE(w, flusher, "done", map[string]interface{}{
		"content":   result.Content,
		"usage":     result.TotalUsage,
		"steps":     result.Steps,
		"thread_id": h.ThreadID,
	})
}

func (h *StreamHandler) runStreaming(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, input string, history []Message) (*StreamRunResult, error) {
	messages := h.agent.buildMessages(input, history)
	tools := h.agent.tools.ToLLMTools()

	model := h.agent.config.Model
	if h.Model != "" {
		model = h.Model
	}
	model = NormalizeModelID(model)

	var totalUsage Usage
	var content strings.Builder
	steps := 0

	for {
		if err := ctx.Err(); err != nil {
			return h.partialResult(content.String(), messages, totalUsage, steps), ErrStreamCancelled
		}

		steps++
		if steps > h.agent.config.MaxSteps {
			return nil, fmt.Errorf("exceeded max steps (%d)", h.agent.config.MaxSteps)
		}

		params := CompletionParams{
			Model:       model,
			Messages:    messages,
			Temperature: h.agent.config.Temperature,
			MaxTokens:   h.agent.config.MaxTokens,
		}
		if len(tools) > 0 {
			params.Tools = tools
		}

		iter, err := h.agent.provider.Stream(ctx, params)
		if err != nil {
			if ctx.Err() != nil {
				return h.partialResult(content.String(), messages, totalUsage, steps), ErrStreamCancelled
			}
			return nil, fmt.Errorf("step %d: %w", steps, err)
		}

		assembled, usage, err := h.consumeStream(ctx, w, flusher, iter, steps)
		if err != nil {
			if ctx.Err() != nil {
				return h.partialResult(content.String(), messages, totalUsage, steps), ErrStreamCancelled
			}
			return nil, fmt.Errorf("step %d stream: %w", steps, err)
		}

		content.WriteString(assembled.Content)
		totalUsage.PromptTokens += usage.PromptTokens
		totalUsage.CompletionTokens += usage.CompletionTokens
		totalUsage.TotalTokens += usage.TotalTokens

		messages = append(messages, *assembled)

		if len(assembled.ToolCalls) == 0 {
			return &StreamRunResult{
				Content:    assembled.Content,
				Messages:   messages,
				TotalUsage: totalUsage,
				Steps:      steps,
			}, nil
		}

		h.writeSSE(w, flusher, "tool_calls", assembled.ToolCalls)

		toolMessages := h.agent.executeToolCalls(ctx, assembled.ToolCalls)
		messages = append(messages, toolMessages...)

		for _, tm := range toolMessages {
			if ctx.Err() != nil {
				return h.partialResult(content.String(), messages, totalUsage, steps), ErrStreamCancelled
			}
			h.writeSSE(w, flusher, "tool_result", map[string]string{
				"tool_call_id": tm.ToolCallID,
				"content":      tm.Content,
			})
		}
	}
}

func (h *StreamHandler) partialResult(content string, messages []Message, usage Usage, steps int) *StreamRunResult {
	return &StreamRunResult{
		Content:    content,
		Messages:   messages,
		TotalUsage: usage,
		Steps:      steps,
	}
}

func (h *StreamHandler) consumeStream(ctx context.Context, w http.ResponseWriter, flusher http.Flusher, iter *StreamIterator, step int) (*Message, Usage, error) {
	defer iter.Close()

	var content strings.Builder
	var toolCalls []ToolCall
	var usage Usage
	toolCallArgs := make(map[int]*strings.Builder)

	for {
		select {
		case <-ctx.Done():
			msg := &Message{
				Role:      RoleAssistant,
				Content:   content.String(),
				ToolCalls: toolCalls,
			}
			return msg, usage, ctx.Err()
		default:
		}

		event, ok := iter.Next()
		if !ok {
			break
		}

		switch event.Type {
		case EventChunk:
			chunk := event.Chunk
			if chunk.Usage != nil {
				usage = *chunk.Usage
			}
			if len(chunk.Choices) == 0 {
				continue
			}

			delta := chunk.Choices[0].Delta
			if delta.Content != "" {
				content.WriteString(delta.Content)
				h.writeSSE(w, flusher, "content", delta.Content)
			}

			for i, tc := range delta.ToolCalls {
				if tc.ID != "" {
					toolCalls = append(toolCalls, ToolCall{ID: tc.ID, Type: tc.Type, Function: FunctionCall{Name: tc.Function.Name}})
					toolCallArgs[len(toolCalls)-1] = &strings.Builder{}
				}
				idx := len(toolCalls) - 1 + i - len(delta.ToolCalls) + 1
				if idx >= 0 && idx < len(toolCalls) {
					if builder, ok := toolCallArgs[idx]; ok {
						builder.WriteString(tc.Function.Arguments)
					}
				}
			}

		case EventError:
			return nil, usage, event.Err

		case EventDone:
			// Stream complete
		}
	}

	for i, builder := range toolCallArgs {
		if i < len(toolCalls) {
			toolCalls[i].Function.Arguments = builder.String()
		}
	}

	msg := &Message{
		Role:      RoleAssistant,
		Content:   content.String(),
		ToolCalls: toolCalls,
	}
	return msg, usage, nil
}

func (h *StreamHandler) writeSSE(w http.ResponseWriter, flusher http.Flusher, event string, data interface{}) {
	b, _ := json.Marshal(data)
	payload := string(b)

	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)
	flusher.Flush()
}
