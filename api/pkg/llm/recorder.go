package llm

import (
	"context"
	"encoding/json"
	"sync"
)

// RecordedCall captures a single tool invocation during an agent run.
type RecordedCall struct {
	Tool   string
	Args   json.RawMessage
	Result json.RawMessage
	Err    error
}

// ToolCallRecorder accumulates tool executions made during an agent run.
// Attach it to a registry via WrapWithRecorder, then inspect Calls() after Run().
type ToolCallRecorder struct {
	mu    sync.Mutex
	calls []RecordedCall
}

// Calls returns a snapshot of every recorded invocation (safe to call after Run).
func (r *ToolCallRecorder) Calls() []RecordedCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]RecordedCall, len(r.calls))
	copy(out, r.calls)
	return out
}

// CallsFor returns only the invocations targeting the named tool.
func (r *ToolCallRecorder) CallsFor(tool string) []RecordedCall {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []RecordedCall
	for _, c := range r.calls {
		if c.Tool == tool {
			out = append(out, c)
		}
	}
	return out
}

// Reset clears all recorded calls between runs.
func (r *ToolCallRecorder) Reset() {
	r.mu.Lock()
	r.calls = r.calls[:0]
	r.mu.Unlock()
}

// WrapWithRecorder returns a new ToolRegistry where every handler is wrapped to
// record calls into rec before dispatching. The original registry is not modified.
func WrapWithRecorder(base *ToolRegistry, rec *ToolCallRecorder) *ToolRegistry {
	wrapped := NewToolRegistry()
	for _, tool := range base.List() {
		orig := tool.Handler
		name := tool.Name
		tool.Handler = func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			result, err := orig(ctx, args)
			rec.mu.Lock()
			rec.calls = append(rec.calls, RecordedCall{
				Tool:   name,
				Args:   args,
				Result: result,
				Err:    err,
			})
			rec.mu.Unlock()
			return result, err
		}
		wrapped.Register(tool) //nolint:errcheck — name and handler are always valid here
	}
	return wrapped
}
