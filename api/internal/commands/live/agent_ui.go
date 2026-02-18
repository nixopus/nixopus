package live

import (
	"context"
	"fmt"
)

// AgentUI is the main orchestrator. It listens on an EventBus and routes events to the writer.
type AgentUI struct {
	ctx    context.Context
	writer *Writer
	bus    *EventBus
}

// NewAgentUI creates the agent UI wired to the given event bus.
func NewAgentUI(bus *EventBus, term *Terminal) *AgentUI {
	w := NewWriter(term)
	return &AgentUI{
		writer: w,
		bus:    bus,
	}
}

// Run is the main loop. It processes events until ctx is cancelled.
// Priority events are checked first via a nested select to ensure
// they are never starved by a flood of normal events.
func (a *AgentUI) Run(ctx context.Context) error {
	a.ctx = ctx

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
		case <-ctx.Done():
			return nil
		}
	}
}

// routeEvent dispatches an event to direct output.
func (a *AgentUI) routeEvent(ev Event) {
	if ev.NeedsLLM {
		a.handleAgentEventFallback(ev)
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
			a.writer.Detail(fmt.Sprintf("[%s] %s", p.StageId, p.Message))
		} else if ev.Message != "" {
			a.writer.Detail(ev.Message)
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
	case EventApprovalNeeded:
		if p, ok := ev.Payload.(ApprovalNeededPayload); ok {
			a.writer.FinishProgress("") // Clear any in-place progress before showing proposal
			a.writer.ApprovalProposal(p.Summary, p.ValidationScore, p.Suggestions, p.Dockerfile)
		}
	case EventError:
		a.writer.Error(ev.Message)
	default:
		if ev.Message != "" {
			a.writer.Detail(ev.Message)
		}
	}
}

// handleAgentEventFallback displays events directly.
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
