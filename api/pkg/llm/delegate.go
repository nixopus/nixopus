package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

type AgentRegistry struct {
	mu     sync.RWMutex
	agents map[string]*Agent
}

func NewAgentRegistry() *AgentRegistry {
	return &AgentRegistry{agents: make(map[string]*Agent)}
}

func (r *AgentRegistry) Register(name string, agent *Agent) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[name] = agent
}

func (r *AgentRegistry) Get(name string) (*Agent, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agent, ok := r.agents[name]
	return agent, ok
}

func (r *AgentRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	names := make([]string, 0, len(r.agents))
	for name := range r.agents {
		names = append(names, name)
	}
	return names
}

func (r *AgentRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.agents)
}

// DelegateTool returns a tool that lets an agent delegate tasks to sub-agents.
func (r *AgentRegistry) DelegateTool(stepLimits map[string]int) ToolDefinition {
	agentNames := r.List()
	agentEnum := "One of: " + join(agentNames, ", ")

	return ToolDefinition{
		Name:        "delegate",
		Description: "Delegate a task to a specialized sub-agent. Include all relevant context in the task description.",
		Parameters:  json.RawMessage(fmt.Sprintf(`{"type":"object","properties":{"agent":{"type":"string","description":"%s"},"task":{"type":"string","description":"What the agent should do — include all relevant context"}},"required":["agent","task"]}`, agentEnum)),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Agent string `json:"agent"`
				Task  string `json:"task"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			agent, ok := r.Get(input.Agent)
			if !ok {
				return json.Marshal(map[string]interface{}{
					"error":     fmt.Sprintf("agent %q not found", input.Agent),
					"available": r.List(),
				})
			}

			if limit, ok := stepLimits[input.Agent]; ok {
				agent = cloneAgentWithMaxSteps(agent, limit)
			}

			result, err := agent.Run(ctx, input.Task)
			if err != nil {
				return json.Marshal(map[string]string{
					"error": fmt.Sprintf("agent %q failed: %s", input.Agent, err.Error()),
				})
			}

			return json.Marshal(map[string]interface{}{
				"agent":   input.Agent,
				"content": result.Content,
				"steps":   result.Steps,
				"usage":   result.TotalUsage,
			})
		},
	}
}

func cloneAgentWithMaxSteps(original *Agent, maxSteps int) *Agent {
	cfg := original.config
	cfg.MaxSteps = maxSteps
	return &Agent{
		config:   cfg,
		provider: original.provider,
		tools:    original.tools,
	}
}

// --- Parallel Execution ---

type ParallelTask struct {
	Agent *Agent
	Input string
}

type ParallelResult struct {
	Index  int
	Result *RunResult
	Err    error
}

// RunParallel executes multiple agent tasks concurrently and returns all results.
func RunParallel(ctx context.Context, tasks []ParallelTask) []ParallelResult {
	results := make([]ParallelResult, len(tasks))
	var wg sync.WaitGroup

	for i, task := range tasks {
		wg.Add(1)
		go func(idx int, t ParallelTask) {
			defer wg.Done()
			result, err := t.Agent.Run(ctx, t.Input)
			results[idx] = ParallelResult{Index: idx, Result: result, Err: err}
		}(i, task)
	}

	wg.Wait()
	return results
}

// ParallelDelegateTool returns a tool that runs multiple sub-agents in parallel.
func (r *AgentRegistry) ParallelDelegateTool() ToolDefinition {
	return ToolDefinition{
		Name:        "parallel_delegate",
		Description: "Run multiple sub-agent tasks in parallel. Use when tasks are independent and can execute concurrently.",
		Parameters:  json.RawMessage(`{"type":"object","properties":{"tasks":{"type":"array","items":{"type":"object","properties":{"agent":{"type":"string"},"task":{"type":"string"}},"required":["agent","task"]},"description":"Array of agent tasks to run in parallel"}},"required":["tasks"]}`),
		Handler: func(ctx context.Context, args json.RawMessage) (json.RawMessage, error) {
			var input struct {
				Tasks []struct {
					Agent string `json:"agent"`
					Task  string `json:"task"`
				} `json:"tasks"`
			}
			if err := json.Unmarshal(args, &input); err != nil {
				return nil, fmt.Errorf("invalid arguments: %w", err)
			}

			tasks := make([]ParallelTask, 0, len(input.Tasks))
			for _, t := range input.Tasks {
				agent, ok := r.Get(t.Agent)
				if !ok {
					return json.Marshal(map[string]interface{}{
						"error":     fmt.Sprintf("agent %q not found", t.Agent),
						"available": r.List(),
					})
				}
				tasks = append(tasks, ParallelTask{Agent: agent, Input: t.Task})
			}

			results := RunParallel(ctx, tasks)

			output := make([]map[string]interface{}, len(results))
			for i, r := range results {
				if r.Err != nil {
					output[i] = map[string]interface{}{
						"agent": input.Tasks[i].Agent,
						"error": r.Err.Error(),
					}
				} else {
					output[i] = map[string]interface{}{
						"agent":   input.Tasks[i].Agent,
						"content": r.Result.Content,
						"steps":   r.Result.Steps,
					}
				}
			}

			return json.Marshal(map[string]interface{}{"results": output})
		},
	}
}

func join(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}
