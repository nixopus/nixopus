package llm

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestPipeline_InputProcessors(t *testing.T) {
	p := NewPipeline()
	p.AddInput(&TokenLimiter{MaxTokens: 1000})

	if p.InputCount() != 1 {
		t.Errorf("expected 1 input processor, got %d", p.InputCount())
	}
}

func TestPipeline_OutputProcessors(t *testing.T) {
	p := NewPipeline()
	p.AddOutput(NewToolBudget(10))

	if p.OutputCount() != 1 {
		t.Errorf("expected 1 output processor, got %d", p.OutputCount())
	}
}

func TestPipeline_RunInput(t *testing.T) {
	p := NewPipeline()
	p.AddInput(NewContextInjector("user_ctx"))

	input := &StepInput{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
		Context:  map[string]interface{}{"user_ctx": "You are org X"},
	}

	err := p.RunInput(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(input.SystemMessages) != 1 {
		t.Errorf("expected 1 system message injected, got %d", len(input.SystemMessages))
	}
}

func TestPipeline_RunInputError(t *testing.T) {
	p := NewPipeline()
	p.AddInput(&failingProcessor{})

	input := &StepInput{Messages: []Message{{Role: RoleUser, Content: "hi"}}}
	err := p.RunInput(context.Background(), input)
	if err == nil {
		t.Fatal("expected error from failing processor")
	}
}

func TestPipeline_RunOutput(t *testing.T) {
	p := NewPipeline()
	tb := NewToolBudget(5)
	p.AddOutput(tb)

	input := &StepInput{Context: make(map[string]interface{})}
	output := &StepOutput{Message: Message{ToolCalls: []ToolCall{{ID: "1"}, {ID: "2"}}}}

	err := p.RunOutput(context.Background(), input, output)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.Context["_tool_budget_used"] != 2 {
		t.Errorf("expected 2 tool calls tracked, got %v", input.Context["_tool_budget_used"])
	}
}

func TestPipeline_RunOutputError(t *testing.T) {
	p := NewPipeline()
	p.AddOutput(&failingOutputProcessor{})

	err := p.RunOutput(context.Background(), &StepInput{}, &StepOutput{})
	if err == nil {
		t.Fatal("expected error from failing output processor")
	}
}

func TestPipeline_Chaining(t *testing.T) {
	p := NewPipeline().
		AddInput(NewTokenLimiter(1000)).
		AddInput(NewContextInjector("ctx")).
		AddOutput(NewToolBudget(10))

	if p.InputCount() != 2 {
		t.Errorf("expected 2, got %d", p.InputCount())
	}
	if p.OutputCount() != 1 {
		t.Errorf("expected 1, got %d", p.OutputCount())
	}
}

// --- TokenLimiter tests ---

func TestTokenLimiter_TruncatesOldMessages(t *testing.T) {
	limiter := NewTokenLimiter(50) // ~200 chars budget

	messages := make([]Message, 10)
	for i := range messages {
		messages[i] = Message{Role: RoleUser, Content: strings.Repeat("x", 40)}
	}

	input := &StepInput{Messages: messages}
	limiter.ProcessInput(context.Background(), input)

	if len(input.Messages) >= 10 {
		t.Errorf("expected truncation, got %d messages", len(input.Messages))
	}
	// Last message should always be preserved
	if input.Messages[len(input.Messages)-1].Content != strings.Repeat("x", 40) {
		t.Error("expected last message preserved")
	}
}

func TestTokenLimiter_PreservesAllWhenUnderBudget(t *testing.T) {
	limiter := NewTokenLimiter(10000)

	messages := []Message{
		{Role: RoleUser, Content: "short"},
		{Role: RoleAssistant, Content: "reply"},
	}

	input := &StepInput{Messages: messages}
	limiter.ProcessInput(context.Background(), input)

	if len(input.Messages) != 2 {
		t.Errorf("expected 2 messages preserved, got %d", len(input.Messages))
	}
}

func TestTokenLimiter_AccountsForSystemMessages(t *testing.T) {
	limiter := NewTokenLimiter(20) // very tight budget

	input := &StepInput{
		SystemMessages: []Message{{Role: RoleSystem, Content: strings.Repeat("s", 60)}},
		Messages: []Message{
			{Role: RoleUser, Content: strings.Repeat("a", 40)},
			{Role: RoleUser, Content: strings.Repeat("b", 40)},
			{Role: RoleUser, Content: strings.Repeat("c", 40)},
		},
	}
	limiter.ProcessInput(context.Background(), input)

	// System takes all budget (60 chars = 15 tokens > 20 budget), messages should be truncated
	if len(input.Messages) >= 3 {
		t.Errorf("expected truncation due to system message, got %d", len(input.Messages))
	}
}

func TestTokenLimiter_AlwaysKeepsAtLeastOne(t *testing.T) {
	limiter := NewTokenLimiter(1) // nearly zero budget

	input := &StepInput{
		Messages: []Message{{Role: RoleUser, Content: strings.Repeat("x", 1000)}},
	}
	limiter.ProcessInput(context.Background(), input)

	if len(input.Messages) != 1 {
		t.Errorf("expected at least 1 message, got %d", len(input.Messages))
	}
}

// --- ToolBudget tests ---

func TestToolBudget_LimitsTools(t *testing.T) {
	tb := NewToolBudget(3)
	ctx := context.Background()

	input := &StepInput{
		Tools:   []Tool{{Type: "function", Function: ToolFunction{Name: "t1"}}},
		Context: map[string]interface{}{},
	}

	// Simulate 3 tool calls
	tb.ProcessOutput(ctx, input, &StepOutput{Message: Message{ToolCalls: []ToolCall{{ID: "1"}, {ID: "2"}, {ID: "3"}}}})

	// Now on next step, tools should be removed
	tb.ProcessInput(ctx, input)
	if input.Tools != nil {
		t.Error("expected tools to be nil after budget exhausted")
	}
}

func TestToolBudget_AllowsUnderLimit(t *testing.T) {
	tb := NewToolBudget(10)
	ctx := context.Background()

	input := &StepInput{
		Tools:   []Tool{{Type: "function", Function: ToolFunction{Name: "t1"}}},
		Context: map[string]interface{}{},
	}

	tb.ProcessOutput(ctx, input, &StepOutput{Message: Message{ToolCalls: []ToolCall{{ID: "1"}}}})
	tb.ProcessInput(ctx, input)

	if len(input.Tools) != 1 {
		t.Errorf("expected tools still available, got %d", len(input.Tools))
	}
}

func TestToolBudget_ReadsFromContext(t *testing.T) {
	tb := NewToolBudget(5)
	ctx := context.Background()

	input := &StepInput{
		Tools:   []Tool{{Type: "function", Function: ToolFunction{Name: "t1"}}},
		Context: map[string]interface{}{"_tool_budget_used": 5},
	}

	tb.ProcessInput(ctx, input)
	if input.Tools != nil {
		t.Error("expected tools nil when context shows budget exhausted")
	}
}

// --- ToolResultPruner tests ---

func TestToolResultPruner_PrunesOldResults(t *testing.T) {
	pruner := NewToolResultPruner(50, 2)

	input := &StepInput{
		StepNumber: 3,
		Messages: []Message{
			{Role: RoleTool, Content: strings.Repeat("x", 200), ToolCallID: "old"},
			{Role: RoleAssistant, Content: "response"},
			{Role: RoleTool, Content: strings.Repeat("y", 200), ToolCallID: "recent"},
			{Role: RoleAssistant, Content: "latest"},
		},
	}

	pruner.ProcessInput(context.Background(), input)

	// First tool message (outside keepRecent) should be pruned
	if len(input.Messages[0].Content) > 70 {
		t.Errorf("expected old tool result pruned, got len %d", len(input.Messages[0].Content))
	}
	if !strings.Contains(input.Messages[0].Content, "[truncated]") {
		t.Error("expected truncation marker")
	}

	// Recent tool message should be preserved
	if len(input.Messages[2].Content) != 200 {
		t.Errorf("expected recent tool result preserved, got len %d", len(input.Messages[2].Content))
	}
}

func TestToolResultPruner_SkipsEarlySteps(t *testing.T) {
	pruner := NewToolResultPruner(10, 0)

	input := &StepInput{
		StepNumber: 1,
		Messages:   []Message{{Role: RoleTool, Content: strings.Repeat("x", 200)}},
	}

	pruner.ProcessInput(context.Background(), input)
	if len(input.Messages[0].Content) != 200 {
		t.Error("expected no pruning on early steps")
	}
}

func TestToolResultPruner_SkipsShortResults(t *testing.T) {
	pruner := NewToolResultPruner(100, 0)

	input := &StepInput{
		StepNumber: 5,
		Messages:   []Message{{Role: RoleTool, Content: "short result"}},
	}

	pruner.ProcessInput(context.Background(), input)
	if input.Messages[0].Content != "short result" {
		t.Error("expected short result preserved")
	}
}

func TestToolResultPruner_NoCutoff(t *testing.T) {
	pruner := NewToolResultPruner(10, 100) // keepRecent > message count

	input := &StepInput{
		StepNumber: 5,
		Messages:   []Message{{Role: RoleTool, Content: strings.Repeat("x", 200)}},
	}

	pruner.ProcessInput(context.Background(), input)
	if len(input.Messages[0].Content) != 200 {
		t.Error("expected no pruning when all messages within keepRecent")
	}
}

// --- ContextInjector tests ---

func TestContextInjector_InjectsContext(t *testing.T) {
	ci := NewContextInjector("org_info")

	input := &StepInput{
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
		Context:  map[string]interface{}{"org_info": "Organization: Acme Corp"},
	}

	ci.ProcessInput(context.Background(), input)
	if len(input.SystemMessages) != 1 {
		t.Fatalf("expected 1 system message, got %d", len(input.SystemMessages))
	}
	if input.SystemMessages[0].Content != "Organization: Acme Corp" {
		t.Errorf("unexpected content: %s", input.SystemMessages[0].Content)
	}
}

func TestContextInjector_NilContext(t *testing.T) {
	ci := NewContextInjector("key")
	input := &StepInput{Context: nil}
	ci.ProcessInput(context.Background(), input)
	if len(input.SystemMessages) != 0 {
		t.Error("expected no injection with nil context")
	}
}

func TestContextInjector_MissingKey(t *testing.T) {
	ci := NewContextInjector("missing")
	input := &StepInput{Context: map[string]interface{}{"other": "val"}}
	ci.ProcessInput(context.Background(), input)
	if len(input.SystemMessages) != 0 {
		t.Error("expected no injection with missing key")
	}
}

func TestContextInjector_EmptyString(t *testing.T) {
	ci := NewContextInjector("key")
	input := &StepInput{Context: map[string]interface{}{"key": ""}}
	ci.ProcessInput(context.Background(), input)
	if len(input.SystemMessages) != 0 {
		t.Error("expected no injection with empty string")
	}
}

func TestContextInjector_NonStringValue(t *testing.T) {
	ci := NewContextInjector("key")
	input := &StepInput{Context: map[string]interface{}{"key": 12345}}
	ci.ProcessInput(context.Background(), input)
	if len(input.SystemMessages) != 0 {
		t.Error("expected no injection with non-string value")
	}
}

// --- Helpers ---

type failingProcessor struct{}

func (f *failingProcessor) ProcessInput(ctx context.Context, input *StepInput) error {
	return fmt.Errorf("processor failed")
}

type failingOutputProcessor struct{}

func (f *failingOutputProcessor) ProcessOutput(ctx context.Context, input *StepInput, output *StepOutput) error {
	return fmt.Errorf("output processor failed")
}
