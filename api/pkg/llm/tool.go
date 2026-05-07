package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type ToolHandler func(ctx context.Context, args json.RawMessage) (json.RawMessage, error)

type ToolDefinition struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
	Handler     ToolHandler     `json:"-"`
}

func (td ToolDefinition) ToLLMTool() Tool {
	return Tool{
		Type: "function",
		Function: ToolFunction{
			Name:        td.Name,
			Description: td.Description,
			Parameters:  td.Parameters,
		},
	}
}

type ToolRegistry struct {
	mu    sync.RWMutex
	tools map[string]ToolDefinition
}

func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[string]ToolDefinition)}
}

func (r *ToolRegistry) Register(tool ToolDefinition) error {
	if tool.Name == "" {
		return fmt.Errorf("llm: tool name cannot be empty")
	}
	if tool.Handler == nil {
		return fmt.Errorf("llm: tool %q handler cannot be nil", tool.Name)
	}
	if tool.Parameters == nil {
		tool.Parameters = json.RawMessage(`{"type":"object","properties":{}}`)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.tools[tool.Name]; exists {
		return fmt.Errorf("llm: tool %q already registered", tool.Name)
	}
	r.tools[tool.Name] = tool
	return nil
}

func (r *ToolRegistry) Get(name string) (ToolDefinition, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *ToolRegistry) Execute(ctx context.Context, name string, args json.RawMessage) (json.RawMessage, error) {
	tool, ok := r.Get(name)
	if !ok {
		return nil, fmt.Errorf("llm: tool %q not found", name)
	}
	return tool.Handler(ctx, args)
}

func (r *ToolRegistry) List() []ToolDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, t)
	}
	return list
}

func (r *ToolRegistry) ToLLMTools() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tools := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t.ToLLMTool())
	}
	return tools
}

func (r *ToolRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.tools)
}

func (r *ToolRegistry) Remove(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.tools[name]; !exists {
		return false
	}
	delete(r.tools, name)
	return true
}
