package llm

import (
	"context"
	"encoding/json"
	"fmt"
)

type AgentConfig struct {
	// Model identifier (e.g. "openai/gpt-4.1-mini")
	Model string
	// System prompt defining agent behavior
	SystemPrompt string
	// Maximum number of tool-call loops before forcing a stop (0 = unlimited)
	MaxSteps int
	// LLM completion parameters
	Temperature *float64
	MaxTokens   *int
}

type Agent struct {
	config   AgentConfig
	provider Provider
	tools    *ToolRegistry
}

func NewAgent(provider Provider, tools *ToolRegistry, config AgentConfig) *Agent {
	if config.MaxSteps == 0 {
		config.MaxSteps = 25
	}
	return &Agent{
		config:   config,
		provider: provider,
		tools:    tools,
	}
}

type RunResult struct {
	// Final assistant message content
	Content string
	// All messages in the conversation (including system, user, assistant, tool)
	Messages []Message
	// Total usage across all LLM calls in this run
	TotalUsage Usage
	// Number of LLM calls made
	Steps int
}

// RunOptions allows per-call overrides of agent configuration.
type RunOptions struct {
	// Model overrides AgentConfig.Model for this run. Empty means use default.
	Model string
}

func (a *Agent) Run(ctx context.Context, input string, history ...Message) (*RunResult, error) {
	return a.RunWithOptions(ctx, RunOptions{}, input, history...)
}

func (a *Agent) RunWithOptions(ctx context.Context, opts RunOptions, input string, history ...Message) (*RunResult, error) {
	model := a.config.Model
	if opts.Model != "" {
		model = opts.Model
	}
	model = NormalizeModelID(model)

	messages := a.buildMessages(input, history)
	tools := a.tools.ToLLMTools()

	var totalUsage Usage
	steps := 0

	for {
		steps++
		if steps > a.config.MaxSteps {
			return nil, fmt.Errorf("llm: agent exceeded max steps (%d)", a.config.MaxSteps)
		}

		params := CompletionParams{
			Model:       model,
			Messages:    messages,
			Temperature: a.config.Temperature,
			MaxTokens:   a.config.MaxTokens,
		}
		if len(tools) > 0 {
			params.Tools = tools
		}

		resp, err := a.provider.Complete(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("llm: agent step %d failed: %w", steps, err)
		}

		totalUsage.PromptTokens += resp.Usage.PromptTokens
		totalUsage.CompletionTokens += resp.Usage.CompletionTokens
		totalUsage.TotalTokens += resp.Usage.TotalTokens

		if len(resp.Choices) == 0 {
			return nil, fmt.Errorf("llm: agent step %d returned no choices", steps)
		}

		choice := resp.Choices[0]
		messages = append(messages, choice.Message)

		if choice.FinishReason != "tool_calls" || len(choice.Message.ToolCalls) == 0 {
			if steps < a.config.MaxSteps && looksLikeUnfinishedPlan(choice.Message.Content) {
				messages = append(messages, Message{
					Role:    RoleUser,
					Content: "[system] You described your next action but didn't execute it. Proceed now — call the tool.",
				})
				continue
			}
			return &RunResult{
				Content:    choice.Message.Content,
				Messages:   messages,
				TotalUsage: totalUsage,
				Steps:      steps,
			}, nil
		}

		toolMessages := a.executeToolCalls(ctx, choice.Message.ToolCalls)
		messages = append(messages, toolMessages...)
	}
}

func (a *Agent) executeToolCalls(ctx context.Context, calls []ToolCall) []Message {
	messages := make([]Message, 0, len(calls))

	for _, tc := range calls {
		result, err := a.tools.Execute(ctx, tc.Function.Name, json.RawMessage(tc.Function.Arguments))
		if err != nil {
			errResult, _ := json.Marshal(map[string]string{"error": err.Error()})
			messages = append(messages, Message{
				Role:       RoleTool,
				Content:    string(errResult),
				ToolCallID: tc.ID,
			})
			continue
		}

		messages = append(messages, Message{
			Role:       RoleTool,
			Content:    string(result),
			ToolCallID: tc.ID,
		})
	}

	return messages
}

func (a *Agent) buildMessages(input string, history []Message) []Message {
	messages := make([]Message, 0, len(history)+2)

	if a.config.SystemPrompt != "" {
		messages = append(messages, Message{Role: RoleSystem, Content: a.config.SystemPrompt})
	}

	messages = append(messages, history...)
	messages = append(messages, Message{Role: RoleUser, Content: input})

	return messages
}
