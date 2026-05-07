package llm

import "context"

type StepInput struct {
	Messages       []Message
	SystemMessages []Message
	Tools          []Tool
	StepNumber     int
	Context        map[string]interface{}
}

type StepOutput struct {
	Message Message
	Usage   Usage
	Context map[string]interface{}
}

type InputProcessor interface {
	ProcessInput(ctx context.Context, input *StepInput) error
}

type OutputProcessor interface {
	ProcessOutput(ctx context.Context, input *StepInput, output *StepOutput) error
}

type Pipeline struct {
	input  []InputProcessor
	output []OutputProcessor
}

func NewPipeline() *Pipeline {
	return &Pipeline{}
}

func (p *Pipeline) AddInput(processors ...InputProcessor) *Pipeline {
	p.input = append(p.input, processors...)
	return p
}

func (p *Pipeline) AddOutput(processors ...OutputProcessor) *Pipeline {
	p.output = append(p.output, processors...)
	return p
}

func (p *Pipeline) RunInput(ctx context.Context, input *StepInput) error {
	for _, proc := range p.input {
		if err := proc.ProcessInput(ctx, input); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) RunOutput(ctx context.Context, input *StepInput, output *StepOutput) error {
	for _, proc := range p.output {
		if err := proc.ProcessOutput(ctx, input, output); err != nil {
			return err
		}
	}
	return nil
}

func (p *Pipeline) InputCount() int {
	return len(p.input)
}

func (p *Pipeline) OutputCount() int {
	return len(p.output)
}

// --- Built-in Processors ---

// TokenLimiter truncates message history to fit within a token budget.
// Uses a simple character-based estimation (1 token ≈ 4 chars).
type TokenLimiter struct {
	MaxTokens int
}

func NewTokenLimiter(maxTokens int) *TokenLimiter {
	return &TokenLimiter{MaxTokens: maxTokens}
}

func (t *TokenLimiter) ProcessInput(ctx context.Context, input *StepInput) error {
	budget := t.MaxTokens
	for _, msg := range input.SystemMessages {
		budget -= estimateTokens(msg.Content)
	}

	// Keep messages from newest to oldest until budget exhausted
	kept := make([]Message, 0, len(input.Messages))
	for i := len(input.Messages) - 1; i >= 0; i-- {
		cost := estimateTokens(input.Messages[i].Content)
		if budget-cost < 0 && len(kept) > 0 {
			break
		}
		budget -= cost
		kept = append(kept, input.Messages[i])
	}

	// Reverse to restore chronological order
	for i, j := 0, len(kept)-1; i < j; i, j = i+1, j-1 {
		kept[i], kept[j] = kept[j], kept[i]
	}

	input.Messages = kept
	return nil
}

func estimateTokens(s string) int {
	return (len(s) + 3) / 4
}

// ToolBudget limits total tool calls across all steps.
type ToolBudget struct {
	Max    int
	used   int
	ctxKey string
}

func NewToolBudget(max int) *ToolBudget {
	return &ToolBudget{Max: max, ctxKey: "_tool_budget_used"}
}

func (tb *ToolBudget) ProcessInput(ctx context.Context, input *StepInput) error {
	if v, ok := input.Context[tb.ctxKey]; ok {
		tb.used = v.(int)
	}
	if tb.used >= tb.Max {
		input.Tools = nil
	}
	return nil
}

func (tb *ToolBudget) ProcessOutput(ctx context.Context, input *StepInput, output *StepOutput) error {
	tb.used += len(output.Message.ToolCalls)
	input.Context[tb.ctxKey] = tb.used
	return nil
}

// ToolResultPruner compresses large tool results in older messages.
type ToolResultPruner struct {
	MaxResultLen int
	KeepRecent   int
}

func NewToolResultPruner(maxLen, keepRecent int) *ToolResultPruner {
	return &ToolResultPruner{MaxResultLen: maxLen, KeepRecent: keepRecent}
}

func (p *ToolResultPruner) ProcessInput(ctx context.Context, input *StepInput) error {
	if input.StepNumber < 2 {
		return nil
	}

	cutoff := len(input.Messages) - p.KeepRecent
	if cutoff <= 0 {
		return nil
	}

	for i := 0; i < cutoff; i++ {
		msg := &input.Messages[i]
		if msg.Role == RoleTool && len(msg.Content) > p.MaxResultLen {
			msg.Content = msg.Content[:p.MaxResultLen] + "\n...[truncated]"
		}
	}
	return nil
}

// ContextInjector adds a system message from the request context.
type ContextInjector struct {
	Key string
}

func NewContextInjector(key string) *ContextInjector {
	return &ContextInjector{Key: key}
}

func (ci *ContextInjector) ProcessInput(ctx context.Context, input *StepInput) error {
	if input.Context == nil {
		return nil
	}
	val, ok := input.Context[ci.Key]
	if !ok {
		return nil
	}
	content, ok := val.(string)
	if !ok || content == "" {
		return nil
	}
	input.SystemMessages = append(input.SystemMessages, Message{Role: RoleSystem, Content: content})
	return nil
}
