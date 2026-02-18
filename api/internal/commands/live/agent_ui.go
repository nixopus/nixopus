package live

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
)

// AgentUI is the main orchestrator that replaces Bubble Tea.
// It listens on an EventBus and routes events to either
// the writer (direct status output) or the agent (intelligent responses).
type AgentUI struct {
	ctx      context.Context
	writer   *Writer
	streamer *Streamer
	input    *InputReader
	bus      *EventBus

	// agentClient communicates with the remote Mastra deploy agent.
	// When nil, agent events fall back to direct output.
	agentClient *AgentClient

	// agentResp receives streamed text chunks from background agent calls.
	agentResp chan string

	// agentBusy is true while an agent call is in-flight.
	agentBusy atomic.Bool

	// agentQueue holds events that arrived while an agent call was in-flight.
	agentQueueMu sync.Mutex
	agentQueue   []Event
}

// NewAgentUI creates the agent UI wired to the given event bus.
// The agent client can be set later via SetAgentClient once auth is available.
func NewAgentUI(bus *EventBus, term *Terminal) *AgentUI {
	w := NewWriter(term)
	return &AgentUI{
		writer:    w,
		streamer:  NewStreamer(w),
		input:     NewInputReader(term.IsTTY()),
		bus:       bus,
		agentResp: make(chan string, 64),
	}
}

// SetAgentClient sets the agent HTTP client. Safe to call while Run is active.
func (a *AgentUI) SetAgentClient(client *AgentClient) {
	a.agentClient = client
}

// Run is the main loop. It processes events until ctx is cancelled.
// Priority events are checked first via a nested select to ensure
// they are never starved by a flood of normal events.
func (a *AgentUI) Run(ctx context.Context) error {
	a.ctx = ctx
	a.input.Start(ctx)

	for {
		// First, drain any priority events without blocking.
		select {
		case ev, ok := <-a.bus.Priority():
			if !ok {
				return nil
			}
			a.routeEvent(ev)
			continue
		default:
		}

		// No priority pending — wait on all sources.
		select {
		case ev, ok := <-a.bus.Priority():
			if !ok {
				return nil
			}
			a.routeEvent(ev)
		case ev, ok := <-a.bus.Normal():
			if !ok {
				return nil
			}
			a.routeEvent(ev)
		case chunk, ok := <-a.agentResp:
			if ok {
				a.streamer.WriteChunk(chunk)
			}
		case text, ok := <-a.input.Chan():
			if !ok {
				continue
			}
			a.handleUserInput(text)
		case <-ctx.Done():
			return nil
		}
	}
}

// routeEvent dispatches an event to either the agent or direct output.
func (a *AgentUI) routeEvent(ev Event) {
	if ev.NeedsLLM {
		a.dispatchToAgent(ev)
	} else {
		a.handleDirectEvent(ev)
	}
}

// handleDirectEvent prints a status line directly without calling the agent.
func (a *AgentUI) handleDirectEvent(ev Event) {
	switch ev.Type {
	case EventConnecting:
		a.writer.Progress("Connecting...")
	case EventConnected:
		a.writer.FinishProgress("Connected")
	case EventReconnecting:
		a.writer.Progress("Reconnecting...")
	case EventDisconnected:
		a.writer.Error("Disconnected")
	case EventSyncStart:
		a.writer.Progress("Syncing files...")
	case EventSyncProgress:
		a.writer.Progress(ev.Message)
	case EventSyncDone:
		a.writer.FinishProgress(ev.Message)
	case EventPipelineProgress:
		if p, ok := ev.Payload.(PipelineProgressPayload); ok {
			a.writer.Progress(fmt.Sprintf("[%s] %s", p.StageId, p.Message))
		} else if ev.Message != "" {
			a.writer.Progress(ev.Message)
		}
	case EventBuildStatus:
		if p, ok := ev.Payload.(BuildStatusPayload); ok {
			switch p.Phase {
			case "error":
				a.writer.Error(p.Message)
				if p.Error != "" {
					a.writer.Detail(fmt.Sprintf("  %s", p.Error))
				}
			case "starting", "generating_dockerfile", "building_container":
				a.writer.Progress(p.Message)
			case "dockerfile_ready":
				a.writer.FinishProgress(p.Message)
			default:
				a.writer.Detail(p.Message)
			}
		}
	case EventBuildLog:
		if p, ok := ev.Payload.(BuildLogPayload); ok {
			a.writer.Detail(p.Log)
		} else if ev.Message != "" {
			a.writer.Detail(ev.Message)
		}
	case EventDeploymentStatus:
		if p, ok := ev.Payload.(DeploymentStatusPayload); ok {
			switch p.Status {
			case "building":
				a.writer.Progress("Building...")
			case "deploying":
				a.writer.Progress("Deploying...")
			case "deployed":
				a.writer.FinishProgress("Deployed")
			case "failed":
				a.writer.Error("Deployment failed")
			default:
				a.writer.Detail(fmt.Sprintf("Status: %s", p.Status))
			}
		}
	case EventBuildStarted:
		a.writer.Progress("Building...")
	case EventDeploying:
		a.writer.Progress("Deploying...")
	case EventDeployed:
		a.writer.FinishProgress(ev.Message)
	case EventInfo:
		a.writer.Detail(ev.Message)
	case EventError:
		a.writer.Error(ev.Message)
	default:
		if ev.Message != "" {
			a.writer.Detail(ev.Message)
		}
	}
}

// dispatchToAgent sends an event to the agent in a background goroutine.
// If an agent call is already in-flight, the event is queued and processed
// when the current call finishes. This prevents blocking the UI loop.
func (a *AgentUI) dispatchToAgent(ev Event) {
	if a.agentBusy.Load() {
		a.agentQueueMu.Lock()
		a.agentQueue = append(a.agentQueue, ev)
		a.agentQueueMu.Unlock()
		return
	}

	a.agentBusy.Store(true)
	go func() {
		defer func() {
			a.agentBusy.Store(false)
			a.drainAgentQueue()
		}()
		a.handleAgentEvent(ev)
	}()
}

// drainAgentQueue processes any events that were queued while an agent call was in-flight.
func (a *AgentUI) drainAgentQueue() {
	for {
		a.agentQueueMu.Lock()
		if len(a.agentQueue) == 0 {
			a.agentQueueMu.Unlock()
			return
		}
		ev := a.agentQueue[0]
		a.agentQueue = a.agentQueue[1:]
		a.agentQueueMu.Unlock()

		a.agentBusy.Store(true)
		a.handleAgentEvent(ev)
		a.agentBusy.Store(false)
	}
}

// handleAgentEvent sends event context to the deploy agent and streams the response.
// Falls back to direct output when the agent client is nil or the call fails.
func (a *AgentUI) handleAgentEvent(ev Event) {
	if a.agentClient == nil {
		a.handleAgentEventFallback(ev)
		return
	}

	// Build the message for the agent based on event type
	msg := a.buildAgentMessage(ev)
	if msg == "" {
		a.handleAgentEventFallback(ev)
		return
	}

	messages := []AgentMessage{{Role: "user", Content: msg}}
	ch, err := a.agentClient.Stream(a.ctx, messages)
	if err != nil {
		a.writer.Warning(fmt.Sprintf("Agent unavailable: %v", err))
		a.handleAgentEventFallback(ev)
		return
	}

	// Stream response chunks to the UI
	gotOutput := false
	for chunk := range ch {
		if strings.HasPrefix(chunk, "[error]") {
			a.writer.Error(strings.TrimPrefix(chunk, "[error] "))
			gotOutput = true
			break
		}
		if !gotOutput {
			a.writer.Blank()
			gotOutput = true
		}
		a.agentResp <- chunk
	}
	a.streamer.Flush()
	if gotOutput {
		a.writer.Blank()
	}
}

// buildAgentMessage constructs a prompt for the agent based on the event.
func (a *AgentUI) buildAgentMessage(ev Event) string {
	switch ev.Type {
	case EventAuth:
		return fmt.Sprintf("[system event: auth] %s", ev.Message)
	case EventConfig:
		return fmt.Sprintf("[system event: config] %s", ev.Message)
	case EventEnvCheck:
		msg := fmt.Sprintf("[system event: env_check] %s", ev.Message)
		if p, ok := ev.Payload.(EnvCheckPayload); ok {
			if len(p.Missing) > 0 {
				msg += fmt.Sprintf("\nMissing env vars: %s", strings.Join(p.Missing, ", "))
			}
			if len(p.Found) > 0 {
				msg += fmt.Sprintf("\nFound env vars: %s", strings.Join(p.Found, ", "))
			}
		}
		return msg
	case EventFileDetected:
		return fmt.Sprintf("[system event: file_detected] %s", ev.Message)
	case EventBuildFailed:
		msg := "[system event: build_failed]"
		if ev.Message != "" {
			msg += " " + ev.Message
		}
		if p, ok := ev.Payload.(BuildFailedPayload); ok {
			if p.Error != "" {
				msg += "\nError: " + p.Error
			}
			if len(p.Logs) > 0 {
				// Send last 30 log lines to keep context window manageable
				start := 0
				if len(p.Logs) > 30 {
					start = len(p.Logs) - 30
				}
				msg += "\nBuild logs (last lines):\n" + strings.Join(p.Logs[start:], "\n")
			}
		}
		return msg
	case EventBuildSuccess:
		return "[system event: build_success] Build completed successfully."
	case EventBuildStatus:
		if p, ok := ev.Payload.(BuildStatusPayload); ok && p.Phase == "error" {
			msg := fmt.Sprintf("[system event: build_error] %s", p.Message)
			if p.Error != "" {
				msg += "\nError details: " + p.Error
			}
			return msg
		}
		return ""
	case EventDeployFailed:
		return fmt.Sprintf("[system event: deploy_failed] %s", ev.Message)
	default:
		if ev.Message != "" {
			return fmt.Sprintf("[system event: %s] %s", ev.Type, ev.Message)
		}
		return ""
	}
}

// handleAgentEventFallback displays events directly when the agent is unavailable.
func (a *AgentUI) handleAgentEventFallback(ev Event) {
	switch ev.Type {
	case EventAuth:
		a.writer.Success(ev.Message)
	case EventConfig:
		a.writer.Success(ev.Message)
	case EventEnvCheck:
		if p, ok := ev.Payload.(EnvCheckPayload); ok && len(p.Missing) > 0 {
			a.writer.Error(ev.Message)
			for _, name := range p.Missing {
				a.writer.Detail(fmt.Sprintf("  missing: %s", name))
			}
		} else {
			a.writer.Success(ev.Message)
		}
	case EventFileDetected:
		a.writer.Success(ev.Message)
	case EventBuildFailed:
		a.writer.Error("Build failed")
		if ev.Message != "" {
			a.writer.Detail(ev.Message)
		}
		if p, ok := ev.Payload.(BuildFailedPayload); ok {
			for _, line := range p.Logs {
				a.writer.Detail(line)
			}
		}
	case EventBuildSuccess:
		a.writer.FinishProgress("Build successful")
	case EventBuildStatus:
		if p, ok := ev.Payload.(BuildStatusPayload); ok && p.Phase == "error" {
			a.writer.Error(p.Message)
			if p.Error != "" {
				a.writer.Detail(fmt.Sprintf("  %s", p.Error))
			}
		}
	case EventDeployFailed:
		a.writer.Error("Deploy failed")
		if ev.Message != "" {
			a.writer.Detail(ev.Message)
		}
	default:
		a.handleDirectEvent(ev)
	}
}

// handleUserInput processes text typed by the user.
// When an agent client is available, sends the question to the agent.
func (a *AgentUI) handleUserInput(text string) {
	if text == "" {
		return
	}
	a.writer.UserInput(text)

	if a.agentClient == nil {
		a.writer.Detail("(Agent not connected)")
		return
	}

	// Send to agent in background via dispatch
	a.dispatchToAgent(Event{
		Type:     EventInfo,
		Message:  text,
		NeedsLLM: true,
	})
}
